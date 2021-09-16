package proxy

import (
	"math/rand"
	"sync"
)

// Secrets is a source of secret values
type Secrets interface {
	Secret() string
}

// randomSecrets returns a random unique secret each time
type randomSecrets struct {
	// rand.Rand isn't concurrent-safe, so guard it
	sync.Mutex

	// size of Secrets to generate
	length int

	// bytes that can occur in generated Secrets
	letters []byte

	// random source
	rand *rand.Rand
}

// NewRandomSecrets returns a random secret generator
func NewRandomSecrets(length int, seed int64) Secrets {

	s := &randomSecrets{
		length:  length,
		letters: []byte("abcdefghijklmnopqrstuvwxyz0123456789"),
		rand:    rand.New(rand.NewSource(seed)),
	}

	return s
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

// fixedSecrets an unchanging secret for testing
type fixedSecrets string

// NewFixedSecret returns a Secrets that always returns the same value
func NewFixedSecret(secret string) Secrets {
	return fixedSecrets(secret)
}

func (s fixedSecrets) Secret() string {
	return string(s)
}
