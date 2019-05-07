package kms

type Resource struct {
	URL string
	Parameter string
	IsBase64 bool
	Data []byte
}

type EncryptRequest struct {
	Key string
	*Resource
	TargetURL string
}


type EncryptResponse struct {
	EncryptedData []byte
	EncryptedText string
}

type DecryptRequest struct {
	Key string
	*Resource

}

type DecryptResponse struct {
	Data []byte
	Text string
}