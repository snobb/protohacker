package ciphers

// Ciphers represents ciphers interface
type Cipher interface {
	Do(b byte, args ...byte) byte
	Undo(b byte, args ...byte) byte
}
