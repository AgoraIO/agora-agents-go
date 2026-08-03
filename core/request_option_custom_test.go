package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPoolUsesInternalFixedBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		area     Area
		expected string
	}{
		{
			name:     "US area",
			area:     AreaUS,
			expected: "https://api-test.agora.io/api/conversational-ai-agent",
		},
		{
			name:     "EU area",
			area:     AreaEU,
			expected: "https://api-test.agora.io/api/conversational-ai-agent",
		},
		{
			name:     "AP area",
			area:     AreaAP,
			expected: "https://api-test.agora.io/api/conversational-ai-agent",
		},
		{
			name:     "CN area",
			area:     AreaCN,
			expected: "https://api-test.agora.io/cn/api/conversational-ai-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(internalAPIBaseURLEnv, "https://api-test.agora.io/")

			pool, err := NewPool(tt.area)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, pool.GetCurrentURL())
		})
	}
}

func TestFixedBaseURLDisablesDynamicRouting(t *testing.T) {
	t.Setenv(internalAPIBaseURLEnv, "https://api-test.agora.io")

	pool, err := NewPool(AreaUS)
	require.NoError(t, err)

	resolverCalled := false
	pool.resolver = ResolverFunc(func(context.Context, []string, string) (string, error) {
		resolverCalled = true
		return "", errors.New("resolver should not be called")
	})

	expected := "https://api-test.agora.io/api/conversational-ai-agent"
	for i := 0; i < 3; i++ {
		pool.NextRegion()
		require.NoError(t, pool.SelectBestDomain(context.Background()))
		assert.Equal(t, expected, pool.GetCurrentURL())
	}
	assert.False(t, resolverCalled)
}

func TestNewPoolAcceptsConfiguredBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "custom HTTPS host",
			baseURL:  "https://staging.example.com/",
			expected: "https://staging.example.com/api/conversational-ai-agent",
		},
		{
			name:     "HTTP localhost with port",
			baseURL:  "http://localhost:8080",
			expected: "http://localhost:8080/api/conversational-ai-agent",
		},
		{
			name:     "existing URL components",
			baseURL:  "https://user:password@staging.example.com/gateway?debug=true#section",
			expected: "https://user:password@staging.example.com/gateway/api/conversational-ai-agent?debug=true#section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(internalAPIBaseURLEnv, tt.baseURL)

			pool, err := NewPool(AreaUS)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, pool.GetCurrentURL())
		})
	}
}

func TestNewPoolWithoutInternalBaseURLKeepsRegionalRouting(t *testing.T) {
	t.Setenv(internalAPIBaseURLEnv, "")

	pool, err := NewPool(AreaUS)
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://api-us-west-1.agora.io/api/conversational-ai-agent",
		pool.GetCurrentURL(),
	)

	pool.NextRegion()
	assert.Equal(
		t,
		"https://api-us-east-1.agora.io/api/conversational-ai-agent",
		pool.GetCurrentURL(),
	)
}

func TestAreaRequestOptionUsesInternalFixedBaseURL(t *testing.T) {
	t.Setenv(internalAPIBaseURLEnv, "https://api-test.agora.io")

	options := NewRequestOptions(NewAreaRequestOption(AreaCN))

	assert.Equal(
		t,
		"https://api-test.agora.io/cn/api/conversational-ai-agent",
		options.BaseURL,
	)
}
