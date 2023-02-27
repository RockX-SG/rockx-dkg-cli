package keystore

type KeyStoreV4 struct {
	Crypto      CryptoInfo `json:"crypto"`
	Description string     `json:"description"`
	PubKey      string     `json:"pubkey"`
	Path        string     `json:"path"`
	UUID        string     `json:"uuid"`
	Version     int        `json:"version"`
}

type CryptoInfo struct {
	KDF      KDFInfo      `json:"kdf"`
	Checksum ChecksumInfo `json:"checksum"`
	Cipher   CipherInfo   `json:"cipher"`
}

type KDFInfo struct {
	Function string    `json:"function"`
	Params   KDFParams `json:"params"`
	Message  string    `json:"message"`
}

type KDFParams struct {
	DKLen int    `json:"dklen"`
	N     int    `json:"n"`
	R     int    `json:"r"`
	P     int    `json:"p"`
	Salt  string `json:"salt"`
}

type ChecksumInfo struct {
	Function string   `json:"function"`
	Params   struct{} `json:"params"`
	Message  string   `json:"message"`
}

type CipherInfo struct {
	Function string       `json:"function"`
	Params   CipherParams `json:"params"`
	Message  string       `json:"message"`
}

type CipherParams struct {
	IV string `json:"iv"`
}
