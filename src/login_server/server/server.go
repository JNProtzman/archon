/*
 * Starting point for the login server. Initializes the configuration package and takes care of
 * launching the LOGIN and CHARACTER servers. Also provides top-level functions and other code
 * shared between the two (found in login.go and character.go).
 */
package server

import (
	"errors"
	"fmt"
	"libarchon/encryption"
	"libarchon/util"
	"net"
	"os"
	"sync"
)

type Client struct {
	conn   *net.TCPConn
	ipAddr string

	clientCrypt *encryption.PSOCrypt
	serverCrypt *encryption.PSOCrypt
	recvData    []byte
	recvSize    int
	packetSize  uint16
}

// Create and initialize a new struct to hold client information.
func NewClient(conn *net.TCPConn) (*Client, error) {
	client := new(Client)
	client.conn = conn
	client.ipAddr = conn.RemoteAddr().String()
	client.recvData = make([]byte, 1024)

	client.clientCrypt = encryption.NewCrypt()
	client.serverCrypt = encryption.NewCrypt()
	client.clientCrypt.CreateKeys()
	client.serverCrypt.CreateKeys()

	var err error = nil
	if SendWelcome(client) != 0 {
		err = errors.New("Error sending welcome packet to: " + client.ipAddr)
		client = nil
	}
	return client, err
}

func Start() {
	// Initialize our config singleton from one of two expected file locations.
	fmt.Printf("Loading config file %v...", loginConfigFile)
	err := GetConfig().InitFromFile(loginConfigFile)
	if err != nil {
		path := util.ServerConfigDir + "/" + loginConfigFile
		fmt.Printf("Failed.\nLoading config from %v...", path)
		err = GetConfig().InitFromFile(path)
		if err != nil {
			fmt.Println("Failed.\nPlease check that one of these files exists and restart the server.")
			os.Exit(-1)
		}
	}
	// TODO: Validate that the configuration struct was populated.
	fmt.Printf("Done.\n--Configuration Parameters--\n%v\n\n", GetConfig().String())

	// Create a WaitGroup so that main won't exit until the server threads have exited.
	var wg sync.WaitGroup
	wg.Add(2)
	go StartLogin(&wg)
	go StartCharacter(&wg)
	wg.Wait()
}
