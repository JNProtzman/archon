package launcher

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/dcrodman/archon"
	archdebug "github.com/dcrodman/archon/internal/debug"
	"github.com/dcrodman/archon/internal/server"
	"io"
	"log"
	"net"
	"runtime/debug"
	"time"
)

// frontend implements the concurrent client connection logic.
//
// Data is read from any connected clients and passed to a backend instance, abstracting
// the lower level connection details away from the Backends.
type frontend struct {
	hostname string
	port     string
	backend  server.Backend
}

func newServerFrontend(hostname, port string, backend server.Backend) *frontend {
	return &frontend{
		hostname: hostname,
		port:     port,
		backend:  backend,
	}
}

// StartListening opens a TCP socket for the specified server and enters a blocking loop
// for accepting client connections and dispatching them to the server.
func (f *frontend) StartListening() error {
	socket, err := f.openSocket()
	if err != nil {
		return err
	}

	log.Printf("waiting for %s connections on %v:%v\n", f.backend.Name(), f.hostname, f.port)

	f.startBlockingLoop(socket)
	return nil
}

func (f *frontend) openSocket() (*net.TCPListener, error) {
	hostAddr, err := net.ResolveTCPAddr("tcp", f.hostname+":"+f.port)
	if err != nil {
		return nil, fmt.Errorf("error creating socket: %s", err.Error())
	}

	socket, err := net.ListenTCP("tcp", hostAddr)
	if err != nil {
		return nil, fmt.Errorf("error listening on socket: %s", err.Error())
	}

	return socket, nil
}

// startBlockingLoop implements a connection handling loop that's purely responsible for
// accepting new connections and spinning off goroutines for the Backend to handle them.
func (f *frontend) startBlockingLoop(socket *net.TCPListener) {
	defer log.Println(f.backend.Name() + " exiting")

	for {
		// Poll until we can accept more clients.
		for isServerFull() {
			time.Sleep(time.Second)
		}

		connection, err := socket.AcceptTCP()
		if err != nil {
			archon.Log.Warnf("failed to accept connection: %s", err.Error())
			continue
		}

		// Immediately spin off a goroutine to handle the client so that the main accept
		// loop doesn't get blocked by one client.
		//
		// Note: If there is eventually a need to implement worker pooling rather than spawning
		// new goroutines for each client, this is where it should be implemented.
		go f.acceptClient(connection)
	}
}

func (f *frontend) acceptClient(connection *net.TCPConn) {
	c := server.NewClient(connection)
	c.Extension = f.backend.CreateExtension()

	if err := f.backend.StartSession(c); err != nil {
		archon.Log.Errorf("StartSession() failed for client %s: %s", c.IPAddr(), err)
	}

	// Prevent multiple clients from connecting from the same IP address.
	if globalClientList.has(c) {
		archon.Log.Infof("rejected second %s connection from %s", f.backend.Name(), c.IPAddr())
		_ = connection.Close()
		return
	}

	archon.Log.Infof("accepted %s connection from %s", f.backend.Name(), c.IPAddr())

	globalClientList.add(c)
	f.processPackets(c)
}

// processPackets starts a blocking loop dedicated to reading data sent from
// a game client and only returns once the connection has closed.
func (f *frontend) processPackets(c *server.Client) {
	// Defer so that we catch any panics, disconnect the client, and
	// remove them from the list regardless of the state of the connection.
	defer f.closeConnectionAndRecover(f.backend.Name(), c)

	buffer := make([]byte, 2048)
	var err error

	for {
		buffer, err = f.readNextPacket(c, buffer)

		if err == io.EOF {
			break
		} else if err != nil {
			archon.Log.Warn(err.Error())
			break
		}

		if archdebug.Enabled() {
			size := determinePacketSize(buffer, c.Extension.HeaderSize())
			archdebug.SendClientPacketToAnalyzer(c.Extension.DebugInfo(), buffer, uint16(size))
		}

		if err = f.backend.Handle(c, buffer); err != nil {
			archon.Log.Warn("error in client communication: " + err.Error())
			return
		}
	}
}

func (*frontend) closeConnectionAndRecover(serverName string, c *server.Client) {
	if err := recover(); err != nil {
		archon.Log.Errorf("error in client communication: %s: %s\n%s\n",
			c.IPAddr(), err, debug.Stack())
	}

	if err := c.Close(); err != nil {
		archon.Log.Warnf("failed to close client connection: %s", err)
	}

	globalClientList.remove(c)

	archon.Log.Infof("disconnected %s client %s", serverName, c.IPAddr())
}

// ReadNextPacket is a blocking call that only returns once the client has
// sent the next packet to be processed. The buffer in c.ConnectionState is
// updated with the decrypted packet.
func (f *frontend) readNextPacket(c *server.Client, buffer []byte) ([]byte, error) {
	headerSize := int(c.Extension.HeaderSize())

	// Read and decrypt the packet header.
	if err := f.readDataFromClient(c, headerSize, buffer); err != nil {
		return buffer, err
	}

	c.Extension.Decrypt(buffer[:headerSize], uint32(headerSize))

	packetSize := determinePacketSize(buffer[:2], uint16(headerSize))

	// Grow the client's receive buffer if they send us a packet bigger than its current capacity.
	if packetSize > cap(buffer) {
		newBuf := make([]byte, cap(buffer)+packetSize)
		copy(newBuf, buffer)
		buffer = newBuf
	}

	// Read and decrypt the rest of the packet.
	if err := f.readDataFromClient(c, packetSize-headerSize, buffer[headerSize:]); err != nil {
		return buffer, err
	}

	c.Extension.Decrypt(buffer[headerSize:packetSize], uint32(packetSize-headerSize))

	return buffer, nil
}

func (f *frontend) readDataFromClient(c *server.Client, n int, buffer []byte) error {
	received := 0

	for received < n {
		bytesRead, err := c.Read(buffer[received:n])
		received += bytesRead

		if bytesRead == 0 || err == io.EOF {
			return err
		} else if err != nil {
			return errors.New("socket error (" + c.IPAddr() + ") " + err.Error())
		}
	}

	return nil
}

// Extract the packet length from the first two bytes of data.
func determinePacketSize(data []byte, headerSize uint16) int {
	if len(data) < 2 {
		// Panic since this shouldn't happen unless something's very wrong.
		panic(errors.New("getSize(): data must be at least two bytes"))
	}

	var size uint16
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &size)

	if err != nil {
		archon.Log.Warn("error decoding packet size:", err)
	}

	// The PSO client occasionally sends packets that are longer than their declared
	// size, but are always a multiple of the length of the packet header. Adjust the
	// expected length just in case in order to avoid leaving stray bytes in the buffer.
	size += size % headerSize

	return int(size)
}
