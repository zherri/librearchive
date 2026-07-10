package utils

import (
	"strings"

	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
)

func GeneratePassphrase() (string, error) {
	bip39.SetWordList(wordlists.English)

	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	allWords := strings.Split(mnemonic, " ")

	tenWords := allWords[:10]

	passphrase := strings.Join(tenWords, " ")

	return passphrase, nil
}
