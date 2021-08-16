package aesecb

import (
	"crypto/cipher"
)

type ecbDecrypter struct {
	b         cipher.Block
	blockSize int
}

func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return &ecbDecrypter{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

func (dec *ecbDecrypter) BlockSize() int { return dec.blockSize }

func (dec *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%dec.blockSize != 0 {
		panic("aesecb: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("aesecb: output smaller than input")
	}
	for len(src) > 0 {
		dec.b.Decrypt(dst, src[:dec.blockSize])
		src = src[dec.blockSize:]
		dst = dst[dec.blockSize:]
	}
}
