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

type AddrWeight struct {
	Addr   string
	Weight int
}

type RandomWeightStrategy struct {
	AddrWeights []AddrWeight
}

func (rs RandomWeightStrategy) Pick() string {
	total := 0
	subsets := make([]int, len(rs.AddrWeights))
	for i, aw := range rs.AddrWeights {
		total += aw.Weight
		subsets[i] = total
	}
	i, target := rand.Intn(total), 0
	// TODO: This can be improved using binary search
	for subsets[i] < target {
		i++
	}
	return rs.AddrWeights[i].Addr
}
