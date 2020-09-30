package hasher

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

const minSecretLength = 32

var b64 = base64.StdEncoding.WithPadding(base64.StdPadding)

// Creates a new instance of the Enigma for generating hash
func NewEnigma(secret string) (*Enigma, error) {
	if len(secret) < minSecretLength {
		return nil, fmt.Errorf("Secret is invalid, it should be at least %d charaters", minSecretLength)
	}
	return &Enigma{secret: []byte(secret)}, nil
}

// Hashes and encodes an input
type Enigma struct {
	secret []byte
}

//
func (e *Enigma) Generate(data string) (string, error) {
	h := hmac.New(sha512.New512_256, e.secret[:])
	_, err := h.Write([]byte(data))

	// Probably shouldn't happen, but gotta handle it if it does
	if err != nil {
		return "", err
	}

	sum := h.Sum(nil)

	return b64.EncodeToString(sum), nil
}
