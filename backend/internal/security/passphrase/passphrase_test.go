package passphrase

import (
	"strings"
	"testing"
)

func TestGenerateReturnsTwelveWords(t *testing.T) {
	phrase, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}
	if words := strings.Fields(phrase); len(words) != WordCount {
		t.Fatalf("Generate() returned %d words, want %d", len(words), WordCount)
	}
}
