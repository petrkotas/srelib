package ocm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateQuery(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "internal ID (32 char hex)",
			identifier: "1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p",
			expected:   "(id = '1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p')",
		},
		{
			name:       "external ID (UUID)",
			identifier: "550e8400-e29b-41d4-a716-446655440000",
			expected:   "(external_id = '550e8400-e29b-41d4-a716-446655440000')",
		},
		{
			name:       "display name",
			identifier: "my-cluster",
			expected:   "(display_name like 'my-cluster')",
		},
		{
			name:       "display name with wildcard",
			identifier: "prod-%",
			expected:   "(display_name like 'prod-%')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateQuery(tt.identifier)
			assert.Equal(t, tt.expected, result, "GenerateQuery should return correct query for %s", tt.name)
		})
	}
}
