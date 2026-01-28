package memory_iplist

import (
	"errors"
	"net"
	"sync"
)

var ErrAlreadyExists = errors.New("ip address already exists")
var ErrNotFound = errors.New("ip address not found")
var ErrEmptyIp = errors.New("ip address is empty")

type MemoryIPList struct {
	mutex sync.Mutex
	data  map[string]struct{}
}

func New() *MemoryIPList {
	return &MemoryIPList{data: make(map[string]struct{})}
}

func (m *MemoryIPList) Contains(ip net.IP) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.data[ip.String()]
	return ok
}

func (m *MemoryIPList) Add(ip string) error {
	if ip == "" {
		return ErrEmptyIp
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.data[ip]; ok {
		return ErrAlreadyExists
	}
	m.data[ip] = struct{}{}

	return nil
}

func (m *MemoryIPList) Remove(ip string) error {
	if ip == "" {
		return ErrEmptyIp
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.data[ip]; ok {
		delete(m.data, ip)
		return nil
	}

	return ErrNotFound
}
