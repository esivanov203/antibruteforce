package limiter

import (
	"sync"
	"time"

	"github.com/esivanov203/antibruteforce/internal/model"
)

type MemoryLimiter struct {
	buckets   map[string]*model.Bucket // хранилище бакетов
	mutex     sync.Mutex
	limit     int           // макс. кол-во токенов в бакете
	period    time.Duration // время полного восст-ия макс. кол-во токенов в бакете
	gcTicker  *time.Ticker  // период запуска горутины очистки неактивных бакетов
	bucketTTL time.Duration // время жизни неактивного бакета
}

func NewMemoryLimiter(limit int, period time.Duration) *MemoryLimiter {
	ml := &MemoryLimiter{
		buckets:   make(map[string]*model.Bucket),
		limit:     limit,
		period:    period,
		bucketTTL: period * 2,
		gcTicker:  time.NewTicker(period),
	}
	go ml.startGC()
	return ml
}

// Allow проверяет бакет по ключу,
// если ключа нет - то создает.
func (ml *MemoryLimiter) Allow(key string) bool {
	ml.mutex.Lock()
	b, exists := ml.buckets[key]
	if !exists {
		b = model.NewBucket(ml.limit, ml.period)
		ml.buckets[key] = b
	}
	ml.mutex.Unlock()

	return b.Allow()
}

// Reset очищает бакет по ключу.
func (ml *MemoryLimiter) Reset(key string) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()
	delete(ml.buckets, key)
}

// startGC запускает периодическую очистку неактивных бакетов.
func (ml *MemoryLimiter) startGC() {
	for range ml.gcTicker.C {
		now := time.Now()
		ml.mutex.Lock()
		for k, b := range ml.buckets {
			b.Mutex.Lock()
			if now.Sub(b.Last) > ml.bucketTTL {
				delete(ml.buckets, k)
			}
			b.Mutex.Unlock()
		}
		ml.mutex.Unlock()
	}
}

// StopGC останавливает очистку (при завершении сервиса).
func (ml *MemoryLimiter) StopGC() {
	if ml.gcTicker != nil {
		ml.gcTicker.Stop()
	}
}
