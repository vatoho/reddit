package idgenerator

import "math/rand"

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type RandomIDGenerator struct{}

func (r *RandomIDGenerator) GenerateID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type IDGenerator interface {
	GenerateID(n int) string
}

type TestIDGenerator struct{}

func (t *TestIDGenerator) GenerateID(_ int) string {
	return "generated_id"
}
