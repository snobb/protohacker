package ciphers

import "math/bits"

// ReverseBits
// reversebits: Reverse the order of bits in the byte, so the least-significant bit becomes the
// most-significant bit, the 2nd-least-significant becomes the 2nd-most-significant, and so on.
type ReverseBits struct{}

func (r ReverseBits) Do(b byte, args ...byte) byte {
	return bits.Reverse8(b)
}

func (r ReverseBits) Undo(b byte, args ...byte) byte {
	return bits.Reverse8(b)
}
