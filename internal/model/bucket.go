package model

import (
	"sync"
	"time"
)

type Bucket struct {
	capacity int     // максимальное количество попыток
	token    float64 // оставшееся количество попыток
	rate     float64 // скорость восстановление попыток в секунду (если rate=0,5 то через 2 сек появится одна попытка)
	Last     time.Time
	Mutex    sync.Mutex
}

// NewBucket
// limit - количество попыток за период - duration,
// например 10 попыток за 1 минуту.
func NewBucket(limit int, duration time.Duration) *Bucket {
	return &Bucket{
		capacity: limit,
		token:    float64(limit),
		rate:     float64(limit) / duration.Seconds(), // токены в секунду
		Last:     time.Now(),
	}
}

func (b *Bucket) Allow() bool {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.Last).Seconds() // время от предыдущей попытки
	b.Last = now

	// восстановление токенов
	b.token += elapsed * b.rate
	if b.token > float64(b.capacity) {
		b.token = float64(b.capacity) // количество попыток не должны превышать максимум
	}

	// если есть токены то используем 1 токен для реализации попытки
	if b.token >= 1 {
		b.token--
		return true
	}

	return false
}
