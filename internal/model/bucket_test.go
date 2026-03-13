package model

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
)

func TestBucketAllow(t *testing.T) {
	t.Run("allow and exceeded", func(t *testing.T) {
		bucket := NewBucket(2, time.Minute, clock.New())
		require.True(t, bucket.Allow())
		require.True(t, bucket.Allow())
		require.False(t, bucket.Allow())
	})

	t.Run("token recovery", func(t *testing.T) {
		clk := clock.NewMock()
		bucket := NewBucket(1, time.Minute, clk)

		require.True(t, bucket.Allow())
		require.False(t, bucket.Allow())

		clk.Add(30 * time.Second)
		require.False(t, bucket.Allow())

		clk.Add(30 * time.Second)
		require.True(t, bucket.Allow())
	})

	t.Run("bucket does not exceed max capacity", func(t *testing.T) {
		clk := clock.NewMock()
		bucket := NewBucket(2, time.Minute, clk)

		require.True(t, bucket.Allow())
		clk.Add(2 * time.Minute)         // должно восстановиться до max = 2 токена
		require.True(t, bucket.Allow())  // -1 токен
		require.True(t, bucket.Allow())  // -1 токен
		require.False(t, bucket.Allow()) // осталось 0 токенов => при восстановлении max capacity не превышается
	})

	t.Run("concurrent access", func(t *testing.T) {
		clk := clock.NewMock()
		bucket := NewBucket(100, time.Second, clk)

		var success atomic.Int64
		wg := sync.WaitGroup{}

		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if bucket.Allow() {
					success.Add(1)
				}
			}()
		}

		wg.Wait()
		// токенов не осталось
		require.False(t, bucket.Allow())
		// при 1000 параллельных запросах успешных = capacity
		require.Equal(t, int64(100), success.Load())
		require.False(t, bucket.Allow())
	})

	t.Run("token recovery partially", func(t *testing.T) {
		mockedClk := clock.NewMock()
		bucket := NewBucket(6, time.Minute, mockedClk)

		for i := 0; i < 6; i++ {
			require.True(t, bucket.Allow())
		}
		// токены закончились
		require.False(t, bucket.Allow())

		// через 10 секунд 1 токен восстановился
		mockedClk.Add(time.Second * 11)
		require.True(t, bucket.Allow())
		require.False(t, bucket.Allow())
		// через 20 секунд 2 токена восстановились
		mockedClk.Add(time.Second * 21)
		require.True(t, bucket.Allow())
		require.True(t, bucket.Allow())
		require.False(t, bucket.Allow())
	})
}
