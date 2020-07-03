package common

import (
	"math/rand"
	"time"
)

type Card struct {
	Value string
	Suit  string
}

type Pairs struct {
	one Card
	two Card
}

type Player struct {
	Hand []Card
	Pars []Pairs
}

func Shuffle(slice []Card) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
		randIdx := r.Intn(n)
		slice[n-1], slice[randIdx] = slice[randIdx], slice[n-1]
	}
}
