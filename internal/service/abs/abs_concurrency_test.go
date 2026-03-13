package abs

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrencyABS(t *testing.T) {
	// initial
	whitelist := &IPListMock{}
	blacklist := &IPListMock{}
	loginLimiter := &LimiterMock{}
	passwordLimiter := &LimiterMock{}
	ipLimiter := &LimiterMock{}

	abs := NewAntiBruteService(loginLimiter, passwordLimiter, ipLimiter, whitelist, blacklist)

	// mocks
	ip := net.ParseIP("1.1.1.1")
	whitelist.On("Contains", ip).Return(false)
	blacklist.On("Contains", ip).Return(false)
	loginLimiter.On("Allow", "login:user").Return(true)
	passwordLimiter.On("Allow", "password:pass").Return(true)
	ipLimiter.On("Allow", "ip:1.1.1.1").Return(true)

	// run 100 concurrent requests
	const numRequests = 100
	var wg sync.WaitGroup
	errChan := make(chan error, numRequests)
	okChan := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := abs.Auth("user", "pass", "1.1.1.1")
			errChan <- err
			okChan <- ok
		}()
	}

	wg.Wait()
	close(errChan)
	close(okChan)

	// Check all results
	successCount := 0
	for err := range errChan {
		require.NoError(t, err)
	}
	for ok := range okChan {
		if ok {
			successCount++
		}
	}

	require.Equal(t, numRequests, successCount)
}
