package ciphers

type Cipher interface {
	Do(b byte, args ...byte) byte
	Undo(b byte, args ...byte) byte
}
