package cron

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_Register(t *testing.T) {
	executor, err := New()
	require.NoError(t, err)

	rt := func(ctx Context, args map[string]interface{}) error {
		return nil
	}

	// first time registration should not error
	err = executor.Register("RT", rt)
	assert.NoError(t, err)

	// re-registration should
	err = executor.Register("RT", rt)
	assert.Error(t, err)
}

func TestExecutor_Add(t *testing.T) {
	executor, err := New()
	require.NoError(t, err)
	err = executor.Add(Entry{})
	assert.NoError(t, err)
}

func TestExecutor_Stop(t *testing.T) {
	executor, err := New()
	require.NoError(t, err)

	go func() {
		err = executor.Start()
		require.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	assert.True(t, executor.IsRunning())
	assert.NotPanics(t, func() { executor.Stop() })
	time.Sleep(100 * time.Millisecond)
	assert.False(t, executor.IsRunning())
}

func TestExecutor_StopAll(t *testing.T) {
	executor, err := New()
	require.NoError(t, err)

	go func() {
		err = executor.Start()
		require.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	assert.True(t, executor.IsRunning())
	assert.NotPanics(t, func() { executor.StopAll() })
	time.Sleep(100 * time.Millisecond)
	assert.False(t, executor.IsRunning())
}

func TestExecutor_CancelAll(t *testing.T) {
	t.Skip()
	executor, err := New()
	require.NoError(t, err)

	go func() {
		err = executor.Start()
		require.NoError(t, err)
	}()

	time.Sleep(time.Millisecond)

	rt := func(ctx Context, args map[string]interface{}) error {
		return nil
	}

	// first time registration should not error
	err = executor.Register("RT", rt)
	require.NoError(t, err)

	err = executor.Add(Entry{Routine: "RT", ID: "ID", Expression: "@every 3s"})
	require.NoError(t, err)

	assert.NotPanics(t, func() { executor.CancelAll() })
}

func TestExecutor_Remove(t *testing.T) {
	executor, err := New()
	require.NoError(t, err)
	err = executor.Add(Entry{ID: "ID"})
	require.NoError(t, err)

	err = executor.Remove("ID")
	assert.NoError(t, err)
}

func TestExecutor_Cancel(t *testing.T) {
	executor, err := New()
	require.NoError(t, err)

	canceled := make(chan bool)
	started := make(chan struct{})
	rt := func(ctx Context, args map[string]interface{}) error {
		// signals the job has started
		started <- struct{}{}

		for {
			select {
			case <-ctx.Done():
				canceled <- true
				return ctx.Err()
			default:
				runtime.Gosched()
			}
		}
	}

	// first time registration should not error
	err = executor.Register("RT", rt)
	require.NoError(t, err)

	err = executor.Add(Entry{ID: "ID", Routine: "RT", Expression: "@every 1s"})
	require.NoError(t, err)

	go func() {
		err = executor.Start()
		require.NoError(t, err)
	}()

	<-started // waiting for the job to start
	require.NoError(t, executor.Cancel("ID"))
	assert.True(t, <-canceled)
}
