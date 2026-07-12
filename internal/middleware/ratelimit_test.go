package middleware

import (
	"testing"
)

func TestMatchRoute(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match success",
			pattern: "/api/v1/users",
			path:    "/api/v1/users",
			want:    true,
		},
		{
			name:    "exact match failure",
			pattern: "/api/v1/users",
			path:    "/api/v1/projects",
			want:    false,
		},
		{
			name:    "wildcard prefix match success",
			pattern: "/api/v1/*",
			path:    "/api/v1/projects/123/keys",
			want:    true,
		},
		{
			name:    "wildcard prefix match failure",
			pattern: "/api/v2/*",
			path:    "/api/v1/users",
			want:    false,
		},
		{
			name:    "all match wildcard",
			pattern: "*",
			path:    "/anything/else/here",
			want:    true,
		},
		{
			name:    "empty pattern",
			pattern: "",
			path:    "/anything",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchRoute(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("matchRoute() = %v, want %v for pattern %q and path %q", got, tt.want, tt.pattern, tt.path)
			}
		})
	}
}
