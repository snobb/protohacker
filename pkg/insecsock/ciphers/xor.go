package ciphers

import "log"

// Xor
// xor(N): XOR the byte by the value N. Note that 0 is a valid value for N.
type Xor struct {
	N byte
}

func (x Xor) Do(b byte, args ...byte) byte {
	return b ^ x.N

}

func (x Xor) Undo(b byte, args ...byte) byte {
	return x.Do(b)
}

// XorPos
// xorpos: XOR the byte by its position in the stream, starting from 0.
type XorPos struct{}

func (x XorPos) Do(b byte, args ...byte) byte {
	if len(args) == 0 {
		log.Print("XorPos: empty pos")
		return b
	}

	return b ^ args[0]

}

func (x XorPos) Undo(b byte, args ...byte) byte {
	return x.Do(b, args...)
}
