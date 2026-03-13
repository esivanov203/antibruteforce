package abs

import (
	"net"

	"github.com/esivanov203/antibruteforce/internal/service"
)

type AntiBruteService struct {
	loginLimiter    service.Limiter
	passwordLimiter service.Limiter
	ipLimiter       service.Limiter

	whitelist service.IPList
	blacklist service.IPList
}

func NewAntiBruteService(
	loginLimiter service.Limiter,
	passwordLimiter service.Limiter,
	ipLimiter service.Limiter,
	whitelist service.IPList,
	blacklist service.IPList,
) *AntiBruteService {
	return &AntiBruteService{
		loginLimiter:    loginLimiter,
		passwordLimiter: passwordLimiter,
		ipLimiter:       ipLimiter,
		whitelist:       whitelist,
		blacklist:       blacklist,
	}
}

func (s *AntiBruteService) Auth(username, password, ip string) (bool, error) {
	IP := net.ParseIP(ip)
	if IP == nil {
		return false, service.ErrInvalidIP
	}
	if username == "" {
		return false, service.ErrEmptyUsername
	}
	if password == "" {
		return false, service.ErrEmptyPassword
	}

	// white list
	if s.whitelist.Contains(IP) {
		return true, nil
	}

	// black list
	if s.blacklist.Contains(IP) {
		return false, nil
	}

	// login rate limit
	if !s.loginLimiter.Allow("login:" + username) {
		return false, nil
	}

	// password rate limit
	if !s.passwordLimiter.Allow("password:" + password) {
		return false, nil
	}

	// ip rate limit
	if !s.ipLimiter.Allow("ip:" + ip) {
		return false, nil
	}

	return true, nil
}

func (s *AntiBruteService) ResetBuckets(login, ip string) error {
	IP := net.ParseIP(ip)
	if IP == nil {
		return service.ErrInvalidIP
	}
	if login == "" {
		return service.ErrEmptyUsername
	}

	s.loginLimiter.Reset("login:" + login)
	s.ipLimiter.Reset("ip:" + ip)

	return nil
}

func (s *AntiBruteService) AddToWhitelist(subnet string) error {
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	return s.whitelist.Add(subnet)
}

func (s *AntiBruteService) RemoveFromWhitelist(subnet string) error {
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	return s.whitelist.Remove(subnet)
}

func (s *AntiBruteService) AddToBlacklist(subnet string) error {
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	return s.blacklist.Add(subnet)
}

func (s *AntiBruteService) RemoveFromBlacklist(subnet string) error {
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	return s.blacklist.Remove(subnet)
}
