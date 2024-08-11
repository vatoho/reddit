package hasher

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetHashPassword(password string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(password))
	if err != nil {
		return "", err
	}
	hashBytes := hash.Sum(nil)
	hashPass := hex.EncodeToString(hashBytes)
	return hashPass, nil
}
