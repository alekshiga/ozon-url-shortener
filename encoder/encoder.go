package encoder

import (
	"crypto/rand"
	"math/big"
)

const (
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	codeLength = 10
)

type Encoder struct{}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode() (string, error) {
	result := make([]byte, codeLength)

	for i := range codeLength {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[randomIndex.Int64()]
	}
	return string(result), nil
}
