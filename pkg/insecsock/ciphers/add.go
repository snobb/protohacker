package ciphers

import (
	"log"
)

// Add - add(N): Add N to the byte, modulo 256. Note that 0 is a valid value for N, and addition
// wraps, so that 255+1=0, 255+2=1, and so on.
type Add struct {
	N byte
}

// Do encodes the byte
func (a Add) Do(b byte, args ...byte) byte {
	return b + a.N

}

// Undo decodes the byte
func (a Add) Undo(b byte, args ...byte) byte {
	return b - a.N
}

// AddPos: Add the position in the stream to the byte, modulo 256, starting from 0.
// wraps, so that 255+1=0, 255+2=1, and so on.
type AddPos struct{}

// Do encodes the byte
func (a *AddPos) Do(b byte, args ...byte) byte {
	if len(args) == 0 {
		log.Print("AddPos: empty pos")
		return b
	}

	return b + args[0]
}

// Undo decodes the byte
func (a *AddPos) Undo(b byte, args ...byte) byte {
	if len(args) == 0 {
		log.Print("AddPos: empty pos")
		return b
	}

	return b - args[0]
}
