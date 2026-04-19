package util

import (
	"net"
	"testing"
)

type parsedIPMatcher interface {
	MatchParsedIP(net.IP) bool
}

func requireParsedIPMatcher(tb testing.TB, tree *CIDRTree) parsedIPMatcher {
	tb.Helper()

	matcher, ok := any(tree).(parsedIPMatcher)
	if !ok {
		tb.Fatalf("CIDRTree must implement MatchParsedIP(net.IP) bool")
	}

	return matcher
}

func mustParseIP(tb testing.TB, raw string) net.IP {
	tb.Helper()

	parsedIP := net.ParseIP(raw)
	if parsedIP == nil {
		tb.Fatalf("failed to parse IP %q", raw)
	}

	return parsedIP
}

func TestCIDRTreeMatchParsedIP(t *testing.T) {
	tree := NewCIDRTree()
	tree.AddCIDR("2600:1f13:a0d:a700::/56")
	tree.AddCIDR("192.168.1.0/24")
	tree.AddCIDR("4.145.74.52/30")

	matcher := requireParsedIPMatcher(t, tree)

	tests := []struct {
		name string
		ip   net.IP
		want bool
	}{
		{
			name: "IPv6 address within CIDR",
			ip:   mustParseIP(t, "2600:1f13:a0d:a700::1"),
			want: true,
		},
		{
			name: "IPv6 address outside CIDR",
			ip:   mustParseIP(t, "2600:1f13:b0d:a700::1"),
			want: false,
		},
		{
			name: "IPv4 address within CIDR",
			ip:   mustParseIP(t, "192.168.1.100"),
			want: true,
		},
		{
			name: "IPv4 address outside CIDR",
			ip:   mustParseIP(t, "192.168.2.1"),
			want: false,
		},
		{
			name: "IPv4-mapped IPv6 matches IPv4 tree",
			ip:   mustParseIP(t, "::ffff:192.168.1.1"),
			want: true,
		},
		{
			name: "Nil IP does not match",
			ip:   nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher.MatchParsedIP(tt.ip)
			if got != tt.want {
				t.Fatalf("MatchParsedIP(%v) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestCIDRTreeMatchParsedIPHasNoAllocs(t *testing.T) {
	tree := NewCIDRTree()
	for i := 0; i < 100; i++ {
		tree.AddCIDR("10.50.0.0/24")
	}

	matcher := requireParsedIPMatcher(t, tree)
	parsedIP := mustParseIP(t, "10.50.0.1")

	var matched bool
	allocs := testing.AllocsPerRun(1000, func() {
		matched = matcher.MatchParsedIP(parsedIP)
	})

	if !matched {
		t.Fatal("expected MatchParsedIP to return true")
	}

	if allocs != 0 {
		t.Fatalf("MatchParsedIP should not allocate, got %.2f allocs/run", allocs)
	}
}

func BenchmarkCIDRTree_Match_String_IPv4(b *testing.B) {
	tree := NewCIDRTree()
	for i := 0; i < 100; i++ {
		tree.AddCIDR("10.50.0.0/24")
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !tree.Match("10.50.0.1") {
			b.Fatal("expected tree.Match to return true")
		}
	}
}

func BenchmarkCIDRTree_MatchParsedIP_IPv4(b *testing.B) {
	tree := NewCIDRTree()
	for i := 0; i < 100; i++ {
		tree.AddCIDR("10.50.0.0/24")
	}

	matcher := requireParsedIPMatcher(b, tree)
	parsedIP := mustParseIP(b, "10.50.0.1")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !matcher.MatchParsedIP(parsedIP) {
			b.Fatal("expected MatchParsedIP to return true")
		}
	}
}

func BenchmarkCIDRTree_Match_String_IPv6(b *testing.B) {
	tree := NewCIDRTree()
	for i := 0; i < 100; i++ {
		tree.AddCIDR("2001:db8:32::/48")
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !tree.Match("2001:db8:32::1") {
			b.Fatal("expected tree.Match to return true")
		}
	}
}

func BenchmarkCIDRTree_MatchParsedIP_IPv6(b *testing.B) {
	tree := NewCIDRTree()
	for i := 0; i < 100; i++ {
		tree.AddCIDR("2001:db8:32::/48")
	}

	matcher := requireParsedIPMatcher(b, tree)
	parsedIP := mustParseIP(b, "2001:db8:32::1")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !matcher.MatchParsedIP(parsedIP) {
			b.Fatal("expected MatchParsedIP to return true")
		}
	}
}
