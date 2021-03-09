package prs

func (p *prsCompressor) Decompress(sizeOnly bool) {
	p.controlByte = p.readByte()
	for {
		if p.readBit() == 1 {
			p.copyByte(!sizeOnly)
			continue
		}
		var length, offset int
		if p.readBit() == 1 {
			offset = p.readShort()
			if offset == 0 {
				break
			}

			length = offset & 0b111
			offset >>= 3
			offset |= -0x2000

			if length == 0 {
				length = int(p.readByte())
				length += 1
			} else {
				length += 2
			}
		} else {
			length = p.readBit()
			length <<= 1
			length |= p.readBit()
			length += 2

			offset = int(p.readByte())
			offset |= -0x100
		}

		for length > 0 {
			p.copyByteAt(offset, !sizeOnly)
			length--
		}
	}
}

func (p *prsCompressor) readBit() int {
	if p.bitPos >= 8 {
		p.controlByte = p.readByte()
		p.bitPos = 0
	}
	flag := int(p.controlByte & 1)
	p.controlByte >>= 1
	p.bitPos++
	return flag
}

func (p *prsCompressor) readByte() byte {
	defer func() {
		p.srcPos++
	}()
	return p.src[p.srcPos]
}

func (p *prsCompressor) readShort() int {
	defer func() {
		p.srcPos += 2
	}()
	return int(p.src[p.srcPos]) + int(p.src[p.srcPos+1])<<8
}

func (p *prsCompressor) copyByte(copy bool) {
	if copy {
		p.dst[p.dstPos] = p.src[p.srcPos]
	}
	p.srcPos++
	p.dstPos++
}

func (p *prsCompressor) copyByteAt(offset int, copy bool) {
	if copy {
		p.dst[p.dstPos] = p.dst[p.dstPos+offset]
	}
	p.dstPos++
}
