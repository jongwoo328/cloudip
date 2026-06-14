package common

import "time"

const AppName = "cloudip"

const DefaultUpdateCheckTTL = 24 * time.Hour

type CloudIpFlag struct {
	Delimiter string
	Format    string
	Header    bool
	NoUpdate  bool
	Verbose   bool
}

type UpdatePolicy struct {
	NoUpdate bool
	TTL      time.Duration
}

func DefaultUpdatePolicy() UpdatePolicy {
	return UpdatePolicy{
		TTL: DefaultUpdateCheckTTL,
	}
}

func (policy UpdatePolicy) EffectiveTTL() time.Duration {
	if policy.TTL <= 0 {
		return DefaultUpdateCheckTTL
	}
	return policy.TTL
}

func ShouldCheckUpdate(lastChecked, now time.Time, ttl time.Duration) bool {
	if lastChecked.IsZero() || ttl <= 0 {
		return true
	}
	return !now.Before(lastChecked.Add(ttl))
}
