package auth

import (
	"errors"
	"testing"
)

func TestParseBearer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		err   error
	}{
		{name: "bearer", input: "Bearer abc.def", want: "abc.def"},
		{name: "case insensitive", input: "bearer token", want: "token"},
		{name: "raw token", input: "token", want: "token"},
		{name: "empty", input: " ", err: ErrTokenRequired},
		{name: "bearer only", input: "Bearer", err: ErrTokenRequired},
		{name: "malformed", input: "Basic token", err: ErrTokenMalformed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBearer(tt.input)
			if !errors.Is(err, tt.err) {
				t.Fatalf("ParseBearer() error = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Fatalf("ParseBearer() = %q, want %q", got, tt.want)
			}
		})
	}
}
