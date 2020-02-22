package main

import (
	"math/rand"
	"sync"
)

type Secrets interface {
	Secret() string
}
type randomSecrets struct {
	// rand.Rand isn't thread safe, so guard it
	sync.Mutex

	// size of secrets to generate
	length int

	// bytes that can occur in generates secrets
	letters []byte

	// random source
	rand *rand.Rand
}

func (s *randomSecrets) Secret() string {
	defer s.Unlock()
	s.Lock()
	b := make([]byte, s.length)
	for i := range b {
		b[i] = s.letters[s.rand.Intn(len(s.letters))]
	}
	return string(b)
}

// Make a Secrets that returns random secret values
func NewRandomSecrets(length int, seed int64) Secrets {

	s := &randomSecrets{
		length:  length,
		letters: []byte("abcdefghijklmnopqrstuvwxyz0123456789"),
		rand:    rand.New(rand.NewSource(seed)),
	}

	return s
}

type fixedSecrets string

// Make a Secrets that always returns the same Secret for easy testing
func NewFixedSecret(secret string) Secrets {
	return fixedSecrets(secret)
}

func (s fixedSecrets) Secret() string {
	return string(s)
}
