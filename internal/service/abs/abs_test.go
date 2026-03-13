package abs

import (
	"net"
	"testing"

	"github.com/esivanov203/antibruteforce/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mocks

type LimiterMock struct {
	mock.Mock
}

func (m *LimiterMock) Allow(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *LimiterMock) Reset(key string) {
	m.Called(key)
}

type IPListMock struct {
	mock.Mock
}

func (m *IPListMock) Contains(ip net.IP) bool {
	args := m.Called(ip)
	return args.Bool(0)
}

func (m *IPListMock) Add(subnet string) error {
	args := m.Called(subnet)
	return args.Error(0)
}

func (m *IPListMock) Remove(subnet string) error {
	args := m.Called(subnet)
	return args.Error(0)
}

// test

func TestAddRemoveWhitelistBlacklist(t *testing.T) {
	whitelist := &IPListMock{}
	blacklist := &IPListMock{}

	wlSubnet := "10.0.0.0/8"
	noWlSubnet := "10.0.0.1/8"
	whitelist.On("Add", wlSubnet).Return(nil).Once()
	whitelist.On("Remove", wlSubnet).Return(nil).Once()
	whitelist.On("Remove", noWlSubnet).Return(service.ErrSubnetNotFound).Once()

	blSubnet := "192.168.0.0/16"
	noBlSubnet := "192.168.0.2/16"
	blacklist.On("Add", blSubnet).Return(nil).Once()
	blacklist.On("Remove", blSubnet).Return(nil).Once()
	blacklist.On("Remove", noBlSubnet).Return(service.ErrSubnetNotFound).Once()

	svc := NewAntiBruteService(&LimiterMock{}, &LimiterMock{}, &LimiterMock{}, whitelist, blacklist)

	// subnet format check
	require.Error(t, svc.AddToWhitelist("438943"))
	require.Error(t, svc.RemoveFromWhitelist(""))
	require.Error(t, svc.AddToBlacklist("jdfkdjkf"))
	require.Error(t, svc.RemoveFromBlacklist("193.56.56.34"))

	// WL
	require.NoError(t, svc.AddToWhitelist(wlSubnet))
	require.NoError(t, svc.RemoveFromWhitelist(wlSubnet))
	err := svc.RemoveFromWhitelist(noWlSubnet)
	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrSubnetNotFound)

	// BL
	require.NoError(t, svc.AddToBlacklist(blSubnet))
	require.NoError(t, svc.RemoveFromBlacklist(blSubnet))
	require.Error(t, svc.RemoveFromBlacklist(noBlSubnet))
	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrSubnetNotFound)

	whitelist.AssertExpectations(t)
	blacklist.AssertExpectations(t)
}

func TestResetBuckets(t *testing.T) {
	loginLimiter := &LimiterMock{}
	ipLimiter := &LimiterMock{}

	l := "user"
	ip := "1.1.1.1"
	loginLimiter.On("Reset", "login:"+l).Once()
	ipLimiter.On("Reset", "ip:"+ip).Once()

	abs := NewAntiBruteService(loginLimiter, &LimiterMock{}, ipLimiter, &IPListMock{}, &IPListMock{})

	require.ErrorIs(t, abs.ResetBuckets("bob", ""), service.ErrInvalidIP)
	require.ErrorIs(t, abs.ResetBuckets("bob", "193.45.56."), service.ErrInvalidIP)
	require.ErrorIs(t, abs.ResetBuckets("", "193.45.56.12"), service.ErrEmptyUsername)

	require.NoError(t, abs.ResetBuckets(l, ip), service.ErrEmptyPassword)

	loginLimiter.AssertExpectations(t)
	ipLimiter.AssertExpectations(t)
}

func TestAuth(t *testing.T) {
	type InitialMock struct {
		abs             *AntiBruteService
		whitelist       *IPListMock
		blacklist       *IPListMock
		loginLimiter    *LimiterMock
		passwordLimiter *LimiterMock
		ipLimiter       *LimiterMock
	}

	newInitialMock := func() *InitialMock {
		whitelist := &IPListMock{}
		blacklist := &IPListMock{}
		loginLimiter := &LimiterMock{}
		passwordLimiter := &LimiterMock{}
		ipLimiter := &LimiterMock{}
		abs := NewAntiBruteService(loginLimiter, passwordLimiter, ipLimiter, whitelist, blacklist)
		return &InitialMock{
			abs:             abs,
			whitelist:       whitelist,
			blacklist:       blacklist,
			loginLimiter:    loginLimiter,
			passwordLimiter: passwordLimiter,
			ipLimiter:       ipLimiter,
		}
	}

	t.Run("Invalid IP", func(t *testing.T) {
		im := newInitialMock()

		ok, err := im.abs.Auth("user", "pass", "256.1.1.1")
		require.ErrorIs(t, err, service.ErrInvalidIP)
		require.False(t, ok)

		ok, err = im.abs.Auth("user", "pass", "")
		require.ErrorIs(t, err, service.ErrInvalidIP)
		require.False(t, ok)
	})

	t.Run("Empty username", func(t *testing.T) {
		im := newInitialMock()

		ok, err := im.abs.Auth("", "pass", "1.1.1.1")
		require.ErrorIs(t, err, service.ErrEmptyUsername)
		require.False(t, ok)
	})

	t.Run("Empty password", func(t *testing.T) {
		im := newInitialMock()

		ok, err := im.abs.Auth("user", "", "1.1.1.1")
		require.ErrorIs(t, err, service.ErrEmptyPassword)
		require.False(t, ok)
	})

	t.Run("In Whitelist IP", func(t *testing.T) {
		im := newInitialMock()

		ip := net.ParseIP("1.1.1.1")
		im.whitelist.On("Contains", ip).Return(true).Once()

		ok, err := im.abs.Auth("user", "pass", "1.1.1.1")
		require.NoError(t, err)
		require.True(t, ok)

		im.whitelist.AssertExpectations(t)
	})

	t.Run("In Blacklist IP", func(t *testing.T) {
		im := newInitialMock()

		ip := net.ParseIP("1.1.1.1")
		im.whitelist.On("Contains", ip).Return(false).Once()
		im.blacklist.On("Contains", ip).Return(true).Once()

		ok, err := im.abs.Auth("user", "pass", "1.1.1.1")
		require.NoError(t, err)
		require.False(t, ok)

		im.blacklist.AssertExpectations(t)
	})

	t.Run("Login rate limited", func(t *testing.T) {
		im := newInitialMock()

		ip := net.ParseIP("1.1.1.1")
		im.whitelist.On("Contains", ip).Return(false).Once()
		im.blacklist.On("Contains", ip).Return(false).Once()
		im.loginLimiter.On("Allow", "login:user").Return(false).Once()
		im.passwordLimiter.On("Allow", "password:pass").Return(true)
		im.ipLimiter.On("Allow", "ip:1.1.1.1").Return(true)

		ok, err := im.abs.Auth("user", "pass", "1.1.1.1")
		require.NoError(t, err)
		require.False(t, ok)

		im.loginLimiter.AssertExpectations(t)
	})

	t.Run("Password rate limit", func(t *testing.T) {
		im := newInitialMock()

		ip := net.ParseIP("1.1.1.1")
		im.whitelist.On("Contains", ip).Return(false).Once()
		im.blacklist.On("Contains", ip).Return(false).Once()
		im.loginLimiter.On("Allow", "login:user").Return(true).Once()
		im.passwordLimiter.On("Allow", "password:pass").Return(false).Once()
		im.ipLimiter.On("Allow", "ip:1.1.1.1").Return(true)

		ok, err := im.abs.Auth("user", "pass", "1.1.1.1")
		require.NoError(t, err)
		require.False(t, ok)

		im.passwordLimiter.AssertExpectations(t)
	})

	t.Run("IP rate limit", func(t *testing.T) {
		im := newInitialMock()

		ip := net.ParseIP("1.1.1.1")
		im.whitelist.On("Contains", ip).Return(false).Once()
		im.blacklist.On("Contains", ip).Return(false).Once()
		im.loginLimiter.On("Allow", "login:user").Return(true).Once()
		im.passwordLimiter.On("Allow", "password:pass").Return(true).Once()
		im.ipLimiter.On("Allow", "ip:1.1.1.1").Return(false).Once()

		ok, err := im.abs.Auth("user", "pass", "1.1.1.1")
		require.NoError(t, err)
		require.False(t, ok)

		im.ipLimiter.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {
		im := newInitialMock()

		ip := net.ParseIP("1.1.1.1")
		im.whitelist.On("Contains", ip).Return(false).Once()
		im.blacklist.On("Contains", ip).Return(false).Once()
		im.loginLimiter.On("Allow", "login:user").Return(true).Once()
		im.passwordLimiter.On("Allow", "password:pass").Return(true).Once()
		im.ipLimiter.On("Allow", "ip:1.1.1.1").Return(true).Once()

		ok, err := im.abs.Auth("user", "pass", "1.1.1.1")
		require.NoError(t, err)
		require.True(t, ok)

		im.loginLimiter.AssertExpectations(t)
		im.passwordLimiter.AssertExpectations(t)
		im.ipLimiter.AssertExpectations(t)
	})
}
