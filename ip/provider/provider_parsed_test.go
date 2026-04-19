package provider

import (
	"net"
	"testing"
)

func newParsedTestProvider(tb testing.TB) *BaseProvider {
	tb.Helper()

	bp := NewBaseProvider("ParsedTestProvider", &mockDataManager{}, func(bp *BaseProvider) error { return nil })
	if err := bp.Initialize(); err != nil {
		tb.Fatalf("failed to initialize provider: %v", err)
	}

	bp.AddIPv4Range("192.168.1.0/24")
	bp.AddIPv4Range("10.0.0.0/8")
	bp.AddIPv6Range("2001:db8::/32")

	return bp
}

func mustParseProviderIP(tb testing.TB, raw string) net.IP {
	tb.Helper()

	parsedIP := net.ParseIP(raw)
	if parsedIP == nil {
		tb.Fatalf("failed to parse IP %q", raw)
	}

	return parsedIP
}

func TestBaseProvider_CheckParsedIP(t *testing.T) {
	bp := newParsedTestProvider(t)

	tests := []struct {
		name      string
		ip        net.IP
		wantMatch bool
		wantError bool
	}{
		{
			name:      "IPv4 in range",
			ip:        mustParseProviderIP(t, "192.168.1.100"),
			wantMatch: true,
		},
		{
			name:      "IPv4 outside range",
			ip:        mustParseProviderIP(t, "192.168.2.1"),
			wantMatch: false,
		},
		{
			name:      "IPv6 in range",
			ip:        mustParseProviderIP(t, "2001:db8::1"),
			wantMatch: true,
		},
		{
			name:      "IPv6 outside range",
			ip:        mustParseProviderIP(t, "2001:db9::1"),
			wantMatch: false,
		},
		{
			name:      "IPv4-mapped IPv6 uses IPv4 tree",
			ip:        mustParseProviderIP(t, "::ffff:10.1.1.1"),
			wantMatch: true,
		},
		{
			name:      "Nil IP returns error",
			ip:        nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bp.CheckParsedIP(tt.ip)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CheckParsedIP(%v) returned unexpected error: %v", tt.ip, err)
			}

			if got != tt.wantMatch {
				t.Fatalf("CheckParsedIP(%v) = %v, want %v", tt.ip, got, tt.wantMatch)
			}
		})
	}
}

func TestBaseProvider_CheckParsedIP_MatchesExpectedResults(t *testing.T) {
	bp := newParsedTestProvider(t)

	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "IPv4 in range", ip: "192.168.1.100", want: true},
		{name: "IPv4 outside range", ip: "192.168.2.1", want: false},
		{name: "IPv4 large range", ip: "10.1.1.1", want: true},
		{name: "IPv6 in range", ip: "2001:db8::1", want: true},
		{name: "IPv6 outside range", ip: "2001:db9::1", want: false},
		{name: "IPv4-mapped IPv6", ip: "::ffff:192.168.1.1", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bp.CheckParsedIP(mustParseProviderIP(t, tt.ip))
			if err != nil {
				t.Fatalf("CheckParsedIP(%q) returned unexpected error: %v", tt.ip, err)
			}

			if got != tt.want {
				t.Fatalf("CheckParsedIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestBaseProvider_CheckParsedIPHasNoAllocs(t *testing.T) {
	bp := newParsedTestProvider(t)
	parsedIP := mustParseProviderIP(t, "10.1.1.1")

	var matched bool
	var err error
	allocs := testing.AllocsPerRun(1000, func() {
		matched, err = bp.CheckParsedIP(parsedIP)
	})

	if err != nil {
		t.Fatalf("CheckParsedIP returned unexpected error: %v", err)
	}

	if !matched {
		t.Fatal("expected CheckParsedIP to return true")
	}

	if allocs != 0 {
		t.Fatalf("CheckParsedIP should not allocate, got %.2f allocs/run", allocs)
	}
}

func BenchmarkBaseProvider_CheckParsedIP_IPv4(b *testing.B) {
	bp := newParsedTestProvider(b)
	parsedIP := mustParseProviderIP(b, "10.1.1.1")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		matched, err := bp.CheckParsedIP(parsedIP)
		if err != nil {
			b.Fatalf("CheckParsedIP returned unexpected error: %v", err)
		}
		if !matched {
			b.Fatal("expected CheckParsedIP to return true")
		}
	}
}

func BenchmarkBaseProvider_CheckParsedIP_IPv6(b *testing.B) {
	bp := newParsedTestProvider(b)
	parsedIP := mustParseProviderIP(b, "2001:db8::1")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		matched, err := bp.CheckParsedIP(parsedIP)
		if err != nil {
			b.Fatalf("CheckParsedIP returned unexpected error: %v", err)
		}
		if !matched {
			b.Fatal("expected CheckParsedIP to return true")
		}
	}
}
