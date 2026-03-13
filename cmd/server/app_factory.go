package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/esivanov203/antibruteforce/internal/service"
	"github.com/esivanov203/antibruteforce/internal/service/abs"
	iplist "github.com/esivanov203/antibruteforce/internal/service/memoryiplist"
	limiter "github.com/esivanov203/antibruteforce/internal/service/memorylimiter"
)

func appFactory(
	dur string,
	loginLimit string,
	passwordLimit string,
	ipLimit string,
) (service.App, error) {
	clk := clock.New()

	var d time.Duration
	switch dur {
	case "second":
		d = time.Second
	case "minute":
		d = time.Minute
	case "hour":
		d = time.Hour
	default:
		return nil, errors.New("time duration must be one of: second, minute, hour")
	}

	ll, e1 := strconv.Atoi(loginLimit)
	pl, e2 := strconv.Atoi(passwordLimit)
	il, e3 := strconv.Atoi(ipLimit)
	if e1 != nil || e2 != nil || e3 != nil {
		return nil, errors.New("loginLimit or passwordLimit or ipLimit are not a number")
	}

	// в этой версии используем единственные на данный момент реализации:
	// memory limiter - хранилище бакетов с соотв-й логикой в ОП
	// memory iplist - хранилище списков с соотв-й логикой в ОП
	loginLimiter := limiter.NewMemoryLimiter(ll, d, clk)
	pwdLimiter := limiter.NewMemoryLimiter(pl, d, clk)
	ipLimiter := limiter.NewMemoryLimiter(il, d, clk)
	wList := iplist.New()
	bList := iplist.New()

	// Cервис также имеет свой интерфейс
	app := abs.NewAntiBruteService(
		loginLimiter,
		pwdLimiter,
		ipLimiter,
		wList,
		bList,
	)
	return app, nil
}
