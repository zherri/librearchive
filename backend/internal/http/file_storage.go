package http

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
)

func randomFilename(extension string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes) + extension, nil
}
func copyFile(path string, source io.Reader) error {
	destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
