package cred

type Encryptor interface {
	Encrypt(src []byte) []byte
}

type Decryptor interface {
	Decrypt(src []byte) []byte
}

type Cipher interface {
	Encryptor
	Decryptor
}
