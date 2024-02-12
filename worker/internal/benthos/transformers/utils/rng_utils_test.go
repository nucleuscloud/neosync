package transformer_utils

import (
	"testing"

	"github.com/zeebo/assert"
)

func Test_getRandomizerNoSeed(t *testing.T) {
	// test that a non-nil *rand.Rand is returned with no seed provided

	randomizerNoSeed := GetRandomizer(nil)
	assert.NotNil(t, randomizerNoSeed)
}

func Test_getRandomizerSeedValue(t *testing.T) {
	// test that a non-nil *rand.Rand is returned when a seed provided

	seed := int64(12)

	randomizerNoSeed := GetRandomizer(&seed)
	assert.NotNil(t, randomizerNoSeed)

}

func Test_geRandomizerSameSeedValueDeterministic(t *testing.T) {

	// test that the randomizer is deterministic and returns the same order of values

	seed := int64(12)

	randomizerNoSeedFirstValue := GetRandomizer(&seed).Intn(100)
	randomizerNoSeedSecondValue := GetRandomizer(&seed).Intn(100)

	if randomizerNoSeedSecondValue != randomizerNoSeedSecondValue {
		t.Errorf("Expected deterministic results with same seed, got %v and %v", randomizerNoSeedFirstValue, randomizerNoSeedSecondValue)
	}
}

func Test_geRandomizerDifferentSeedsValue(t *testing.T) {

	// test that the randomizer produces different values for different seeds

	seed := int64(12)

	randomizerNoSeedFirstValue := GetRandomizer(&seed).Intn(30)
	randomizerNoSeedSecondValue := GetRandomizer(&seed).Intn(100)

	if randomizerNoSeedSecondValue == randomizerNoSeedSecondValue {
		t.Errorf("Expected deterministic results with same seed, got %v and %v", randomizerNoSeedFirstValue, randomizerNoSeedSecondValue)
	}
}
