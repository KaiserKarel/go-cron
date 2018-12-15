package cron_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cron "github.com/kaiserkarel/go-cron"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_3s(t *testing.T) {
	if testing.Short() {
		t.Skip("short flag enabled, skipping e2e TestExecutor")
	}

	executor, err := cron.New()
	require.NoError(t, err)

	watchchan := make(chan string)
	testfunc := func(ctx context.Context, args map[string]interface{}) error {
		fmt.Println("called", time.Now())
		watchchan <- args["test"].(string)
		return nil
	}

	executor.Register("test", testfunc)

	executor.Add(cron.Entry{
		ID:         "testjob",
		Expression: "@every 3s",
		Routine:    "test",
		Args: map[string]interface{}{
			"test": "test",
		},
	})

	started := time.Now()
	go executor.Start()
	defer executor.StopAll()

	for i := 0; i < 4; i++ {
		assert.Equal(t, "test", <-watchchan)
	}
	elapsed := time.Now().Sub(started)
	assert.True(t, elapsed > (9*time.Second), elapsed.String())
}

func TestExecutor_1s(t *testing.T) {
	if testing.Short() {
		t.Skip("short flag enabled, skipping e2e TestExecutor")
	}

	executor, err := cron.New()
	require.NoError(t, err)

	watchchan := make(chan string)
	testfunc := func(ctx context.Context, args map[string]interface{}) error {
		fmt.Println("called", time.Now())
		watchchan <- args["test"].(string)
		return nil
	}

	executor.Register("test", testfunc)

	executor.Add(cron.Entry{
		ID:         "testjob",
		Expression: "@every 1s",
		Routine:    "test",
		Args: map[string]interface{}{
			"test": "test",
		},
	})

	started := time.Now()
	go executor.Start()

	for i := 0; i < 4; i++ {
		assert.Equal(t, "test", <-watchchan)
	}
	elapsed := time.Now().Sub(started)
	assert.True(t, elapsed > (3*time.Second), elapsed.String())
}
