package memoryiplist

import (
	"errors"
	"net"
	"sync"
)

var (
	ErrAlreadyExists = errors.New("subnet already exists")
	ErrNotFound      = errors.New("subnet not found")
	ErrEmptyIP       = errors.New("subnet is empty")
)

type MemoryIPList struct {
	mutex sync.Mutex
	data  []*net.IPNet
}

func New() *MemoryIPList {
	return &MemoryIPList{}
}

// проверяет, попадает ли IP в любую подсеть списка.
func (m *MemoryIPList) Contains(ip net.IP) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, val := range m.data {
		if val.Contains(ip) {
			return true
		}
	}

	return false
}

func (m *MemoryIPList) Add(subnet string) error {
	if subnet == "" {
		return ErrEmptyIP
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, n := range m.data {
		if n.String() == ipNet.String() {
			return ErrAlreadyExists
		}
	}

	m.data = append(m.data, ipNet)
	return nil
}

func (m *MemoryIPList) Remove(subnet string) error {
	if subnet == "" {
		return ErrEmptyIP
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, n := range m.data {
		if n.String() == ipNet.String() {
			m.data = append(m.data[:i], m.data[i+1:]...)
			return nil
		}
	}

	return ErrNotFound
}
