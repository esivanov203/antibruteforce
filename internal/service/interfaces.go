package service

import (
	"errors"
	"net"
)

var (
	ErrInvalidIP      = errors.New("invalid ip address")
	ErrSubnetNotFound = errors.New("subnet format check fault")
	ErrEmptyUsername  = errors.New("empty username")
	ErrEmptyPassword  = errors.New("empty password")
)

type App interface {
	Auth(username, password, ip string) (bool, error)
	ResetBuckets(login, ip string) error

	AddToWhitelist(subnet string) error
	RemoveFromWhitelist(subnet string) error
	AddToBlacklist(subnet string) error
	RemoveFromBlacklist(subnet string) error
}

type Limiter interface {
	Allow(key string) bool
	Reset(key string)
}

type IPList interface {
	Contains(ip net.IP) bool
	Add(ip string) error
	Remove(ip string) error
}
