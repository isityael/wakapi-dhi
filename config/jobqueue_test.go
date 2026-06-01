package config

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobQueueDispatchRunsTaskAndUpdatesMetrics(t *testing.T) {
	q := NewJobQueue("test", 1, 8)
	defer q.Stop()

	done := make(chan struct{})
	require.NoError(t, q.Dispatch(func() {
		close(done)
	}))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("queued task did not run")
	}

	require.Eventually(t, func() bool {
		return q.CountDispatched() == 1 && q.CountEnqueued() == 0
	}, time.Second, 10*time.Millisecond)
}

func TestJobQueueDispatchEveryCanBeStopped(t *testing.T) {
	q := NewJobQueue("test", 1, 8)
	defer q.Stop()

	var runs atomic.Int32
	var ranOnce sync.Once
	ran := make(chan struct{})
	ticker, err := q.DispatchEvery(func() {
		runs.Add(1)
		ranOnce.Do(func() {
			close(ran)
		})
	}, 25*time.Millisecond)
	require.NoError(t, err)

	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Fatal("periodic task did not run")
	}

	ticker.Stop()
	afterStop := runs.Load()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, afterStop, runs.Load())
}

func TestJobQueueDispatchCronAcceptsSixFieldExpressions(t *testing.T) {
	q := NewJobQueue("test", 1, 8)
	defer q.Stop()

	cron, err := q.DispatchCron(func() {}, "0 0 18 * * *")
	require.NoError(t, err)
	defer cron.Stop()
}
