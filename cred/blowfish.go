package cred

import (
	"crypto/cipher"
	"golang.org/x/crypto/blowfish"
)

func blowfishChecksizeAndPad(padded []byte) []byte {
	modulus := len(padded) % blowfish.BlockSize
	if modulus != 0 {
		padlen := blowfish.BlockSize - modulus
		for i := 0; i < padlen; i++ {
			padded = append(padded, 0)
		}
	}
	return padded
}

type blowfishCipher struct {
	cipher *blowfish.Cipher
}

func (b *blowfishCipher) Encrypt(source []byte) []byte {
	paddedSource := blowfishChecksizeAndPad(source)
	ciphertext := make([]byte, blowfish.BlockSize+len(paddedSource))
	eiv := ciphertext[:blowfish.BlockSize]
	encodedBlackEncryptor := cipher.NewCBCEncrypter(b.cipher, eiv)
	encodedBlackEncryptor.CryptBlocks(ciphertext[blowfish.BlockSize:], paddedSource)
	return ciphertext
}

func (b *blowfishCipher) Decrypt(encrypted []byte) []byte {
	div := encrypted[:blowfish.BlockSize]
	decrypted := encrypted[blowfish.BlockSize:]
	if len(decrypted)%blowfish.BlockSize != 0 {
		panic("decrypted is not a multiple of blowfish.BlockSize")
	}
	dcbc := cipher.NewCBCDecrypter(b.cipher, div)
	dcbc.CryptBlocks(decrypted, decrypted)
	var result = make([]byte, 0)
	for _, b := range decrypted {
		if b == 0x0 {
			break
		}
		result = append(result, b)
	}
	return result
}

func NewBlowfishCipher(key []byte) (Cipher, error) {
	var passwordCipher, err = blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &blowfishCipher{
		cipher: passwordCipher,
	}, nil
}
