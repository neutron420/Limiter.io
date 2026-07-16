package services

import "testing"

func TestCanRead(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{role: "owner", want: true},
		{role: "admin", want: true},
		{role: "member", want: true},
		{role: "", want: false},
		{role: "unknown", want: true},
	}

	for _, tt := range tests {
		if got := CanRead(tt.role); got != tt.want {
			t.Errorf("CanRead(%q) = %v, want %v", tt.role, got, tt.want)
		}
	}
}

func TestCanWrite(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{role: "owner", want: true},
		{role: "admin", want: true},
		{role: "member", want: false},
		{role: "", want: false},
		{role: "unknown", want: false},
	}

	for _, tt := range tests {
		if got := CanWrite(tt.role); got != tt.want {
			t.Errorf("CanWrite(%q) = %v, want %v", tt.role, got, tt.want)
		}
	}
}
