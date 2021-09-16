package proxy

import (
	"testing"
)

func TestNewFixedSecret(t *testing.T) {
	secrets := NewFixedSecret("abc")
	first := secrets.Secret()
	second := secrets.Secret()
	if first != second {
		t.Fatal("secrets mismatch:", first, second)
	}
}

func TestNewRandomSecrets(t *testing.T) {
	secrets := NewFixedSecret("abc")
	first := secrets.Secret()
	second := secrets.Secret()
	if first != second {
		t.Fatal("secrets match, but shouldn't:", first, second)
	}
}
