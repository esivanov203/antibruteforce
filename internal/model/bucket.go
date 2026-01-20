package model

import (
	"sync"
	"time"
)

type Bucket struct {
	capacity int
	token    float64
	rate     float64
	last     time.Time
	mutex    sync.Mutex
}

func (b *Bucket) Allow() bool {
	return true
}
