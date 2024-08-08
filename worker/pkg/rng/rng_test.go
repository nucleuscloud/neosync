package rng

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_New(t *testing.T) {
	require.NotEmpty(t, New(1))
}

func Test_Int63n(t *testing.T) {
	randomizer := New(1)
	require.GreaterOrEqual(t, randomizer.Int63n(1), int64(0))
}

func Test_Intn(t *testing.T) {
	randomizer := New(1)
	require.GreaterOrEqual(t, randomizer.Intn(1), 0)
}

func Test_Float(t *testing.T) {
	randomizer := New(time.Now().UnixNano())
	require.GreaterOrEqual(t, randomizer.Float64(), float64(0))
}

// Tests that concurrent access is okay as otherwise it would panic
func Test_Parallel(t *testing.T) {
	randomizer := New(1)

	errgrp := errgroup.Group{}
	i := 0
	for i < 1000 {
		errgrp.Go(func() error {
			randomizer.Intn(123)
			return nil
		})
		i++
	}
	err := errgrp.Wait()
	require.NoError(t, err)
}
