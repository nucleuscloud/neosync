package transformer_utils

import "math/rand"

func GetRandomizer(seed *int64) *rand.Rand {

	var randomizer *rand.Rand
	if seed == nil {
		randomizer = rand.New(rand.NewSource(rand.Int63()))
	} else {
		randomizer = rand.New(rand.NewSource(*seed))
	}

	return randomizer
}
