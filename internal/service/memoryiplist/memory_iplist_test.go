package memoryiplist

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryIPList_AddContainsRemove(t *testing.T) {
	list := New()

	ip := "127.0.0.1"
	parsedIP := net.ParseIP(ip)

	// Сначала не должно быть IP
	require.False(t, list.Contains(parsedIP), "IP should not exist yet")

	// Добавляем IP
	require.Error(t, list.Add(""))
	require.NoError(t, list.Add(ip))

	// Теперь должно содержать
	require.True(t, list.Contains(parsedIP), "IP should exist after adding")

	// Добавление повторно → ErrAlreadyExists
	require.ErrorIs(t, list.Add(ip), ErrAlreadyExists)

	// Удаление IP
	require.Error(t, list.Remove(""))
	require.NoError(t, list.Remove(ip))

	// После удаления IP должно быть нет
	require.False(t, list.Contains(parsedIP), "IP should be removed")

	// Удаление несуществующего → ErrNotFound
	require.ErrorIs(t, list.Remove(ip), ErrNotFound)
}

func TestMemoryIPList_Concurrency(t *testing.T) {
	list := New()
	ipBase := "192.168.0."

	var wg sync.WaitGroup
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ip := fmt.Sprintf("%s%d", ipBase, i)
			require.NoError(t, list.Add(ip))
			list.Contains(net.ParseIP(ip))
			require.NoError(t, list.Remove(ip))
		}(i)
	}
	wg.Wait()
}
