package prs

// PRS compression/decompression library. Original underlying C implementation
// by Fuzziqer Software, wrapper written in Go for use with archon.

func Compress(src []byte) []byte {
	p := newPRSCompressor(src, 0)
	p.Compress()
	return p.Dest()
}

func Decompress(src []byte) []byte {
	size := DecompressSize(src)
	p := newPRSCompressor(src, size)
	sizeOnly := false
	p.Decompress(sizeOnly)
	return p.dst
}

func DecompressSize(src []byte) int {
	p := newPRSCompressor(src, 0)
	sizeOnly := true
	p.Decompress(sizeOnly)
	return p.dstPos
}

type prsCompressor struct {
	bitPos           int
	controlByte      byte
	controlByteIndex int

	src    []byte
	dst    []byte
	srcPos int
	dstPos int
}

func newPRSCompressor(src []byte, size int) *prsCompressor {
	if size == 0 {
		size = len(src) * 4
	}
	return &prsCompressor{
		bitPos:      0,
		controlByte: 0,

		srcPos: 0,
		dstPos: 0,
		src:    src,
		dst:    make([]byte, size),
	}
}

func (p *prsCompressor) Dest() []byte {
	return p.dst[:p.dstPos]
}
