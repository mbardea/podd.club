package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Password struct {
	scheme string
	salt   []byte
	hash   []byte
}

const FIELD_SEPARATOR string = ":"

func InitRandom() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// password - cleartext password
// encoded - md5:salt:hash
func CheckPassword(text string, encodedRefPassword string) (bool, error) {
	scheme, salt, err := ParsePassword(encodedRefPassword)
	if err != nil {
		return false, fmt.Errorf("Cannot verify password: %s", err)
	}
	if scheme != "md5" {
		return false, fmt.Errorf("Unsupported password scheme: %s", scheme)
	}

	fmt.Printf("Ref Salt    : %s\n", salt)
	textHash := md5.Sum([]byte(salt + text))
	encodedPassword := EncodePassword(scheme, salt, textHash[:])
	fmt.Printf("Encoded    : %s\n", encodedPassword)
	fmt.Printf("Encoded ref: %s\n", encodedRefPassword)
	return (encodedPassword == encodedRefPassword), nil
}

func MakePassword(text string) string {
	var salt [10]byte
	for i := range salt {
		salt[i] = byte(rand.Uint32())
	}
	stringSalt := hex.EncodeToString(salt[:])

	hash := md5.Sum([]byte(stringSalt + text))
	return EncodePassword("md5", stringSalt, hash[:])
}

func ParsePassword(encoded string) (rscheme string, rsalt string, err error) {
	parts := strings.Split(encoded, FIELD_SEPARATOR)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("Invalid password encoding for '%s'. Expected 3 parts", encoded)
	}
	scheme := parts[0]
	salt := parts[1]
	return scheme, salt, nil
}

func EncodePassword(scheme string, salt string, hash []byte) string {
	return scheme +
		FIELD_SEPARATOR +
		salt +
		FIELD_SEPARATOR +
		hex.EncodeToString(hash)
}
