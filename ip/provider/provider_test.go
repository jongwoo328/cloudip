package provider

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

type mockDataManager struct {
	shouldError bool
	dataURL     string
}

func (m *mockDataManager) EnsureDataFile() error {
	if m.shouldError {
		return errors.New("failed to ensure data file")
	}
	return nil
}

func (m *mockDataManager) GetDataURL() string {
	return m.dataURL
}

func TestNewBaseProvider(t *testing.T) {
	mockDM := &mockDataManager{dataURL: "http://example.com"}
	bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	if bp.name != "TestProvider" {
		t.Errorf("Expected name 'TestProvider', got '%s'", bp.name)
	}
	
	if bp.dataManager != mockDM {
		t.Error("DataManager not set correctly")
	}
	
	if bp.initialized {
		t.Error("Provider should not be initialized by default")
	}
}

func TestBaseProvider_GetName(t *testing.T) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("MyProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	if bp.GetName() != "MyProvider" {
		t.Errorf("Expected name 'MyProvider', got '%s'", bp.GetName())
	}
}

func TestBaseProvider_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		shouldError bool
		expectError bool
	}{
		{
			name:        "Successful initialization",
			shouldError: false,
			expectError: false,
		},
		{
			name:        "Failed initialization",
			shouldError: true,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDM := &mockDataManager{shouldError: tt.shouldError}
			bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
			
			err := bp.Initialize()
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if bp.initialized {
					t.Error("Provider should not be initialized after error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if !bp.initialized {
				t.Error("Provider should be initialized after successful Initialize()")
			}
			
			if bp.v4Tree == nil {
				t.Error("IPv4 tree should be initialized")
			}
			
			if bp.v6Tree == nil {
				t.Error("IPv6 tree should be initialized")
			}
		})
	}
}

func TestBaseProvider_InitializeIdempotent(t *testing.T) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	// First initialization
	err1 := bp.Initialize()
	if err1 != nil {
		t.Fatalf("First initialization failed: %v", err1)
	}
	
	// Second initialization should not error and should be idempotent
	err2 := bp.Initialize()
	if err2 != nil {
		t.Errorf("Second initialization failed: %v", err2)
	}
	
	if !bp.initialized {
		t.Error("Provider should remain initialized")
	}
}

func TestBaseProvider_ConcurrentInitialize(t *testing.T) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	
	// Launch multiple goroutines trying to initialize concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bp.Initialize()
			errors <- err
		}()
	}
	
	wg.Wait()
	close(errors)
	
	// Check that all initializations succeeded
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent initialization failed: %v", err)
		}
	}
	
	if !bp.initialized {
		t.Error("Provider should be initialized after concurrent calls")
	}
}

func TestBaseProvider_CheckIP(t *testing.T) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	// Initialize the provider
	err := bp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}
	
	// Add some test CIDR ranges
	bp.AddIPv4Range("192.168.1.0/24")
	bp.AddIPv4Range("10.0.0.0/8")
	bp.AddIPv6Range("2001:db8::/32")
	
	tests := []struct {
		name        string
		ip          string
		expected    bool
		expectError bool
	}{
		{
			name:        "IPv4 in range",
			ip:          "192.168.1.100",
			expected:    true,
			expectError: false,
		},
		{
			name:        "IPv4 not in range",
			ip:          "192.168.2.1",
			expected:    false,
			expectError: false,
		},
		{
			name:        "IPv4 in large range",
			ip:          "10.1.1.1",
			expected:    true,
			expectError: false,
		},
		{
			name:        "IPv6 in range",
			ip:          "2001:db8::1",
			expected:    true,
			expectError: false,
		},
		{
			name:        "IPv6 not in range",
			ip:          "2001:db9::1",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Invalid IP",
			ip:          "invalid-ip",
			expected:    false,
			expectError: true,
		},
		{
			name:        "Empty IP",
			ip:          "",
			expected:    false,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bp.CheckIP(tt.ip)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for IP %s", tt.expected, result, tt.ip)
			}
		})
	}
}

func TestBaseProvider_AddIPRanges(t *testing.T) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	err := bp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}
	
	// Test AddIPv4Range
	bp.AddIPv4Range("192.168.1.0/24")
	match, err := bp.CheckIP("192.168.1.1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !match {
		t.Error("IPv4 range not added correctly")
	}
	
	// Test AddIPv6Range
	bp.AddIPv6Range("2001:db8::/32")
	match, err = bp.CheckIP("2001:db8::1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !match {
		t.Error("IPv6 range not added correctly")
	}
}

func TestBaseProvider_AddCIDRRange(t *testing.T) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	err := bp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}
	
	tests := []struct {
		name        string
		cidr        string
		testIP      string
		expected    bool
		expectError bool
	}{
		{
			name:        "Valid IPv4 CIDR",
			cidr:        "192.168.1.0/24",
			testIP:      "192.168.1.1",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Valid IPv6 CIDR",
			cidr:        "2001:db8::/32",
			testIP:      "2001:db8::1",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Invalid CIDR",
			cidr:        "invalid-cidr",
			testIP:      "",
			expected:    false,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bp.AddCIDRRange(tt.cidr)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if tt.testIP != "" {
				match, err := bp.CheckIP(tt.testIP)
				if err != nil {
					t.Errorf("Unexpected error checking IP: %v", err)
					return
				}
				
				if match != tt.expected {
					t.Errorf("Expected %v, got %v for IP %s", tt.expected, match, tt.testIP)
				}
			}
		})
	}
}

func TestBaseProvider_IPv4vsIPv6Separation(t *testing.T) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("TestProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	err := bp.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}
	
	// Add IPv4 range
	bp.AddIPv4Range("192.168.1.0/24")
	
	// Add IPv6 range
	bp.AddIPv6Range("2001:db8::/32")
	
	// Test IPv4 doesn't match IPv6 tree and vice versa
	v4Match, err := bp.CheckIP("192.168.1.1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !v4Match {
		t.Error("IPv4 should match in IPv4 range")
	}
	
	v6Match, err := bp.CheckIP("2001:db8::1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !v6Match {
		t.Error("IPv6 should match in IPv6 range")
	}
	
	// Test that IPv4 doesn't match different range
	v4NoMatch, err := bp.CheckIP("10.1.1.1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if v4NoMatch {
		t.Error("IPv4 should not match outside range")
	}
}

func BenchmarkBaseProvider_RepeatedInitialize(b *testing.B) {
	mockDM := &mockDataManager{}
	loadCount := 0
	bp := NewBaseProvider("BenchProvider", mockDM, func(bp *BaseProvider) error {
		loadCount++
		for i := 0; i < 1000; i++ {
			bp.AddIPv4Range(fmt.Sprintf("10.%d.%d.0/24", i/256, i%256))
		}
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.Initialize()
	}
	b.StopTimer()

	if loadCount != 1 {
		b.Errorf("loadFunc called %d times, expected 1", loadCount)
	}
}

func BenchmarkBaseProvider_CheckIP_IPv4(b *testing.B) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("BenchProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	bp.Initialize()
	
	// Add multiple IPv4 ranges
	for i := 0; i < 100; i++ {
		bp.AddIPv4Range(fmt.Sprintf("10.%d.0.0/24", i))
	}
	
	testIP := "10.50.0.1"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bp.CheckIP(testIP)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkBaseProvider_CheckIP_IPv6(b *testing.B) {
	mockDM := &mockDataManager{}
	bp := NewBaseProvider("BenchProvider", mockDM, func(bp *BaseProvider) error { return nil })
	
	bp.Initialize()
	
	// Add multiple IPv6 ranges
	for i := 0; i < 100; i++ {
		bp.AddIPv6Range(fmt.Sprintf("2001:db8:%x::/48", i))
	}
	
	testIP := "2001:db8:32::1"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bp.CheckIP(testIP)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}