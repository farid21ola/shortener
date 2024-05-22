package random

import (
	"math/rand"
	"time"
)

//TODO write tests

func NewRandomString(size int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	alphabet := []rune("qwertyuiopasdfghjklzxcvbnm" +
		"QWERTYUIIOPASDFGHJKLZXCVBNM" +
		"0123456789")

	alias := make([]rune, size)
	for i := range alias {
		alias[i] = alphabet[rnd.Intn(len(alphabet))]
	}

	return string(alias)
}
