package prs

func (p *prsCompressor) Compress() {
	for p.srcPos < len(p.src) {
		var (
			offset int
			length int
			mlen   int
		)
		for y := p.srcPos - 1; (y >= 0 && y >= (p.srcPos-0x1FF0)) && mlen < 256; y-- {
			mlen = 1
			if p.src[y] == p.src[p.srcPos] {
				for mlen <= 256 &&
					p.srcPos+mlen <= len(p.src) &&
					p.src[y+mlen-1] == p.src[p.srcPos+mlen-1] {
					mlen++
				}
				if ((mlen >= 2 && y-p.srcPos >= -0x100) || mlen >= 3) && mlen > length {
					offset = y - p.srcPos
					length = mlen
				}
			}
		}
		if length == 0 {
			p.setBit(1)
			p.copyLiteral()
			continue
		}
		p.copyBlock(offset, length)
		p.srcPos += length
	}
	p.writeEOF()
}

func (p *prsCompressor) setBit(bit int) {
	if p.bitPos-1 == 0 {
		p.dst[p.controlByteIndex] = p.controlByte
		p.controlByteIndex = p.dstPos
		p.dstPos++
		p.controlByte = 0
		p.bitPos = 7
	}
	p.controlByte >>= 1
	if bit != 0 {
		p.controlByte |= 0x80
	}
}

func (p *prsCompressor) copyLiteral() {
	p.dst[p.dstPos] = p.src[p.srcPos]
	p.srcPos++
	p.dstPos++
}

func (p *prsCompressor) copyBlock(offset, length int) {
	if length >= 2 && length <= 5 && offset >= -256 {
		p.setBit(0)
		p.setBit(0)
		p.setBit((length - 2) & 2)
		p.setBit((length - 2) & 1)
		p.writeLiteral(byte(offset))
	} else if length <= 9 {
		p.setBit(0)
		p.setBit(1)
		p.writeLiteral(byte(((offset << 3) & 0xF8) | ((length - 2) & 0x07)))
		p.writeLiteral(byte(offset >> 5))
	} else {
		p.setBit(0)
		p.setBit(1)
		p.writeLiteral((byte)((offset << 3) & 0xF8))
		p.writeLiteral((byte)(offset >> 5))
		p.writeLiteral((byte)(length - 1))
	}
}

func (p *prsCompressor) writeLiteral(literal byte) {
	p.dst[p.dstPos] = literal
	p.dstPos += 1
}

func (p *prsCompressor) writeFinalFlags() {
	p.controlByte >>= p.bitPos
	p.dst[p.controlByteIndex] = p.controlByte
}

func (p *prsCompressor) writeEOF() {
	p.setBit(0)
	p.setBit(1)
	p.writeFinalFlags()
	p.writeLiteral(0)
	p.writeLiteral(0)
}
