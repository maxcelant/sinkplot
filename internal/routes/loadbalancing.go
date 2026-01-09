package routes

import (
	"math/rand"
)

type LoadbalanceStrategy interface {
	Pick() string
}

type RandomStrategy struct {
	Addrs []string
}

func (rs RandomStrategy) Pick() string {
	i := rand.Intn(len(rs.Addrs))
	return rs.Addrs[i]
}
