package model

import (
	"sync"
	"time"

	"github.com/benbjohnson/clock"
)

type Bucket struct {
	capacity int     // максимальное количество попыток
	token    float64 // оставшееся количество попыток
	rate     float64 // скорость восстановление попыток в секунду (если rate=0,5 то через 2 сек появится одна попытка)
	Last     time.Time
	Mutex    sync.Mutex
	clk      clock.Clock
}

// NewBucket
// limit - количество попыток за период - duration,
// например 10 попыток за 1 минуту.
func NewBucket(limit int, duration time.Duration, clk clock.Clock) *Bucket {
	return &Bucket{
		capacity: limit,
		token:    float64(limit),
		rate:     float64(limit) / duration.Seconds(), // токены в секунду
		Last:     clk.Now(),
		clk:      clk,
	}
}

func (b *Bucket) Allow() bool {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	now := b.clk.Now()
	elapsed := now.Sub(b.Last).Seconds()
	b.Last = now

	b.token += elapsed * b.rate
	if b.token > float64(b.capacity) {
		b.token = float64(b.capacity)
	}

	if b.token >= 1 {
		b.token--
		return true
	}

	return false
}
