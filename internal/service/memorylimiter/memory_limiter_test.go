package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
)

func TestMemoryLimiterAllow(t *testing.T) {
	t.Run("simple allow", func(t *testing.T) {
		mockedClk := clock.NewMock()
		m := NewMemoryLimiter(2, time.Minute, mockedClk)

		key1 := "login:user1"
		key2 := "login:user2"
		// успешное добавление
		require.True(t, m.Allow(key1))
		require.True(t, m.Allow(key1))
		// превышение лимита
		require.False(t, m.Allow(key1))

		// новый ключ - новый бакет (со своим лимитом)
		require.True(t, m.Allow(key2))
		require.True(t, m.Allow(key2))
		require.False(t, m.Allow(key2))

		// сброс ключа
		m.Reset(key1)
		// ключ доступен
		require.True(t, m.Allow(key1))
		// другие ключи не сбрасываются
		require.False(t, m.Allow(key2))
		// сброс ключа - полностью восстанавливает лимит
		require.True(t, m.Allow(key1))
		require.False(t, m.Allow(key1))

		// восстановление токенов по времени
		mockedClk.Add(time.Minute + time.Second)
		require.True(t, m.Allow(key1))
	})

	t.Run("GS test", func(t *testing.T) {
		mockedClk := clock.NewMock()
		m := NewMemoryLimiter(1, time.Minute, mockedClk)
		defer m.StopGC()

		key := "pass:strong"

		// создаем бакет
		require.True(t, m.Allow(key))
		// бакет в лимитере присутствует
		_, ok := m.buckets[key]
		require.True(t, ok)
		// прошло времени ttl бакетов = 2 * время восстановления лимита бакета
		mockedClk.Add(5*time.Minute + time.Second)
		// бакета в лимитере нет
		_, ok = m.buckets[key]
		require.False(t, ok)
	})

	t.Run("concurrent access", func(t *testing.T) {
		mockedClk := clock.NewMock()
		m := NewMemoryLimiter(100, time.Second, mockedClk)

		var wg sync.WaitGroup
		var success atomic.Int64

		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if m.Allow("login:user") {
					success.Add(1)
				}
			}()
		}
		wg.Wait()

		require.Equal(t, int64(100), success.Load()) // не более capacity
	})
}
