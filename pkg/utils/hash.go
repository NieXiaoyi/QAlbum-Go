package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func ComputeMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func ComputeSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func ComputeMD5FromBytes(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
