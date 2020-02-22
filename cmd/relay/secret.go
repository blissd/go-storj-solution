package main

import "math/rand"

// length of generated secrets
const secretLength = 6

// bytes that can occur in a secret
var letters = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

func generateSecret(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
