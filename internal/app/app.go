package app

import "net"

type App interface {
	Auth(username, password string, ip net.IP) bool
	ResetBuckets(login string, ip net.IP) error

	AddToWhiteList(subnet string) error
	RemoveFromWhiteList(subnet string) error
	AddToBlackList(subnet string) error
	RemoveFromBlackList(subnet string) error
}
