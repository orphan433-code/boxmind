package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func Generate6Digit() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}

	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

func Hash(code, secret string) string {
	sum := sha256.Sum256([]byte(code + secret))
	return hex.EncodeToString(sum[:])
}
