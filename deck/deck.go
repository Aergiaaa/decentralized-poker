package deck

import (
	"fmt"
	"math/rand"
	"strconv"
)

type Suit int

func (s Suit) String() string {
	switch s {
	case Spades:
		return "spades"
	case Hearts:
		return "hearts"
	case Diamonds:
		return "diamonds"
	case Clubs:
		return "clubs"
	default:
		panic("invalid card suit")
	}

}

func (s Suit) UniCode() string {
	switch s {
	case Spades:
		return "\u2660" // ♠
	case Hearts:
		return "\u2665" // ♥
	case Diamonds:
		return "\u2666" // ♦
	case Clubs:
		return "\u2663" // ♣
	default:
		panic("invalid card suit")
	}

}

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

type Card struct {
	Suit
	Value int
}

func NewCard(s Suit, v int) Card {
	if v > 13 {
		panic("the value of card cannot be higher than 13")
	}

	return Card{
		Suit:  s,
		Value: v,
	}
}

func (c Card) String() string {
	val := strconv.Itoa(c.Value)
	switch c.Value {
	case 1:
		val = "ace"
	case 11:
		val = "jack"
	case 12:
		val = "queen"
	case 13:
		val = "king"
	}

	return fmt.Sprintf("%s of %s %s ", val, c.Suit, c.Suit.UniCode())
}

type Deck [52]Card

func New() Deck {
	var (
		suits = 4
		cards = 13
		d     = [52]Card{}
	)

	index := 0
	for i := range suits {
		for j := range cards {
			d[index] = NewCard(Suit(i), j+1)
			index++
		}
	}

	return shuffle(d)
}

func shuffle(d Deck) Deck {
	r := 0
	for i := range d {
		r = rand.Intn(len(d))

		if r != i {
			d[i], d[r] = d[r], d[i]
		}
	}

	return d
}
