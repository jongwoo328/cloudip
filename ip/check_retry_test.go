package ip

import (
	"cloudip/common"
	"cloudip/ip/provider"
	"errors"
	"testing"
)

func TestCheckBoundsInitializationFailuresAndRetriesNextCall(t *testing.T) {
	failure := errors.New("network unavailable")
	p := &parsedPathMockProvider{name: "AWS", initErr: failure}
	checker := NewIPChecker(map[common.CloudProvider]provider.CloudProvider{common.AWS: p}, []common.CloudProvider{common.AWS})
	results := checker.Check([]string{"bad-ip", "1.2.3.4", "1.2.3.5", "1.2.3.6"})
	if p.initializeCalls != 1 || p.checkParsedCalls != 0 {
		t.Fatalf("init=%d check=%d", p.initializeCalls, p.checkParsedCalls)
	}
	for _, result := range results[1:] {
		if !errors.Is(result.Error, failure) {
			t.Fatalf("missing failure: %+v", result)
		}
	}
	p.initErr = nil
	p.parsedMatch = true
	results = checker.Check([]string{"1.2.3.4"})
	if p.initializeCalls != 2 || results[0].Error != nil || results[0].Provider != common.AWS {
		t.Fatalf("next call did not recover: calls=%d results=%+v", p.initializeCalls, results)
	}
}

func TestCheckContinuesWithHealthyProviderAfterInitializationFailure(t *testing.T) {
	failed := &parsedPathMockProvider{name: "AWS", initErr: errors.New("unavailable")}
	healthy := &parsedPathMockProvider{name: "GCP", parsedMatch: true}
	checker := NewIPChecker(map[common.CloudProvider]provider.CloudProvider{common.AWS: failed, common.GCP: healthy}, []common.CloudProvider{common.AWS, common.GCP})
	results := checker.Check([]string{"1.2.3.4", "1.2.3.5"})
	if failed.initializeCalls != 1 {
		t.Fatalf("failed provider initialized %d times", failed.initializeCalls)
	}
	for _, result := range results {
		if result.Error != nil || result.Provider != common.GCP {
			t.Fatalf("healthy provider did not match: %+v", result)
		}
	}
}
