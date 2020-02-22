package main

import "math/rand"

type Secrets interface {
	Secret() string
}
type randomSecrets struct {
	// size of secrets to generate
	length int

	// bytes that can occur in generates secrets
	letters []byte

	// random source
	r rand.Rand
}

func (s randomSecrets) Secret() string {
	b := make([]byte, s.length)
	for i := range b {
		b[i] = s.letters[rand.Intn(len(s.letters))]
	}
	return string(b)
}

// Make a Secrets that returns random secret values
func NewRandomSecrets(length int) Secrets {
	return randomSecrets{
		length:  length,
		letters: []byte("abcdefghijklmnopqrstuvwxyz0123456789"),
	}
}

type fixedSecrets string

// Make a Secrets that always returns the same Secret for easy testing
func NewFixedSecret(secret string) Secrets {
	return fixedSecrets(secret)
}

func (s fixedSecrets) Secret() string {
	return string(s)
}
