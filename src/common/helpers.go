package common



type Card struct {
	Value string
	Suit string
}

type Pairs struct {
	one Card
	two Card
}

type Player struct {
	Hand []Card
	Pars []Pairs
}