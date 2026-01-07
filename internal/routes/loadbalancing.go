package routes

import "math/rand"

type LoadbalanceStrategy interface {
	Pick([]string) string
}

type RandomStrategy struct{}

func (rs RandomStrategy) Pick(addrs []string) string {
	i := rand.Intn(len(addrs))
	return addrs[i]
}
