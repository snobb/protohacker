package ciphers

import "log"

// Xor - xor(N): XOR the byte by the value N. Note that 0 is a valid value for N.
type Xor struct {
	N byte
}

// Do encodes the byte
func (x Xor) Do(b byte, args ...byte) byte {
	return b ^ x.N

}

// Undo decodes the byte
func (x Xor) Undo(b byte, args ...byte) byte {
	return x.Do(b)
}

// XorPos -  XOR the byte by its position in the stream, starting from 0.
type XorPos struct{}

// Do encodes the byte
func (x XorPos) Do(b byte, args ...byte) byte {
	if len(args) == 0 {
		log.Print("XorPos: empty pos")
		return b
	}

	return b ^ args[0]
}

// Undo decodes the byte
func (x XorPos) Undo(b byte, args ...byte) byte {
	return x.Do(b, args...)
}
