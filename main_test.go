package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRun tests the CLI entry point with various argument combinations.
func TestRun(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args     []string
		wantCode int
	}{
		"no command specified": {
			args:     []string{"cmdgroup"},
			wantCode: gokrazyDoNotSuperviseExitCode,
		},
		"help flag": {
			args:     []string{"cmdgroup", "-help"},
			wantCode: 0,
		},
		"invalid flag": {
			args:     []string{"cmdgroup", "-nonexistent"},
			wantCode: gokrazyDoNotSuperviseExitCode,
		},
		"command not found": {
			args:     []string{"cmdgroup", "/nonexistent/binary"},
			wantCode: gokrazyDoNotSuperviseExitCode,
		},
		"invalid watch value": {
			args:     []string{"cmdgroup", "-watch", "abc", "true"},
			wantCode: gokrazyDoNotSuperviseExitCode,
		},
		"successful command": {
			args:     []string{"cmdgroup", "true"},
			wantCode: 0,
		},
		"failing command": {
			args:     []string{"cmdgroup", "false"},
			wantCode: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			code := run(t.Context(), tt.args)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}
