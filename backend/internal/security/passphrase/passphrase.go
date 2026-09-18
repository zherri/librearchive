package passphrase

import (
	"crypto/sha256"
	"strings"

	bip39 "github.com/decred/go-bip39"
)

const (
	WordCount   = 12
	EntropyBits = 128
)

// Generate creates a BIP-39 English mnemonic from 128 bits of cryptographic entropy.
// BIP-39 appends a checksum, yielding exactly twelve words for this entropy size.
func Generate() (string, error) {
	entropy, err := bip39.NewEntropy(EntropyBits)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// CredentialBytes normalizes whitespace and pre-hashes a mnemonic before bcrypt.
// This avoids bcrypt's 72-byte input limit without discarding any mnemonic words.
func CredentialBytes(phrase string) []byte {
	normalized := strings.Join(strings.Fields(phrase), " ")
	digest := sha256.Sum256([]byte(normalized))
	return digest[:]
}
