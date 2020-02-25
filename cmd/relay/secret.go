package main

import (
	"math/rand"
	"sync"
)

type secrets interface {
	secret() string
}
type randomSecrets struct {
	// rand.Rand isn't concurrent-safe, so guard it
	sync.Mutex

	// size of secrets to generate
	length int

	// bytes that can occur in generated secrets
	letters []byte

	// random source
	rand *rand.Rand
}

// Generates a new random secret.
func (s *randomSecrets) secret() string {
	defer s.Unlock()
	s.Lock()
	b := make([]byte, s.length)
	for i := range b {
		b[i] = s.letters[s.rand.Intn(len(s.letters))]
	}
	return string(b)
}

// Make a Secrets that returns random secret values
func newRandomSecrets(length int, seed int64) secrets {

	s := &randomSecrets{
		length:  length,
		letters: []byte("abcdefghijklmnopqrstuvwxyz0123456789"),
		rand:    rand.New(rand.NewSource(seed)),
	}

	return s
}

type fixedSecrets string

// Make a Secrets that always returns the same secret for easy testing
func newFixedSecret(secret string) secrets {
	return fixedSecrets(secret)
}

func (s fixedSecrets) secret() string {
	return string(s)
}
