package main

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSlicesSplitSeq tests splitting slices around a separator element.
func TestSlicesSplitSeq(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args []string
		want [][]string
	}{
		"nil input": {
			args: nil,
			want: [][]string{nil},
		},
		"empty slice": {
			args: []string{},
			want: [][]string{{}},
		},
		"no separator": {
			args: []string{"arg1"},
			want: [][]string{{"arg1"}},
		},
		"double-dash prefix is not separator": {
			args: []string{"--flag1"},
			want: [][]string{{"--flag1"}},
		},
		"only separator": {
			args: []string{"--"},
			want: [][]string{{}, {}},
		},
		"separator between elements": {
			args: []string{"arg1", "--", "arg2"},
			want: [][]string{{"arg1"}, {"arg2"}},
		},
		"trailing separator": {
			args: []string{"arg1", "--"},
			want: [][]string{{"arg1"}, {}},
		},
		"leading and trailing separator": {
			args: []string{"--", "arg1", "--"},
			want: [][]string{{}, {"arg1"}, {}},
		},
		"multiple separators": {
			args: []string{"arg1", "-flag1", "--", "arg2", "--", "arg3", "-flag3"},
			want: [][]string{{"arg1", "-flag1"}, {"arg2"}, {"arg3", "-flag3"}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := slices.Collect(slicesSplitSeq(tt.args, "--"))
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseInts tests parsing comma-separated integers.
func TestParseInts(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		want    []int
		wantErr assert.ErrorAssertionFunc
	}{
		"empty string": {
			input:   "",
			want:    nil,
			wantErr: assert.NoError,
		},
		"single value": {
			input:   "1",
			want:    []int{1},
			wantErr: assert.NoError,
		},
		"single with whitespace": {
			input:   "  1  ",
			want:    []int{1},
			wantErr: assert.NoError,
		},
		"comma separated": {
			input:   "0, 2",
			want:    []int{0, 2},
			wantErr: assert.NoError,
		},
		"consecutive commas": {
			input:   "0,,2",
			want:    []int{0, 2},
			wantErr: assert.NoError,
		},
		"non-numeric": {
			input:   "a",
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := parseInts(tt.input)
			assert.Equal(t, tt.want, got)
			tt.wantErr(t, err)
		})
	}
}
