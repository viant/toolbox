package url


//AES256Key represents custom key
type AES256Key struct {
	Key                 []byte
	Base64Key           string
	Base64KeyMd5Hash    string
	Base64KeySha256Hash string
}
