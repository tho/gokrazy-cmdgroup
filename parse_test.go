package cmdgroup

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args []string
		want []*Instance
	}{
		{
			args: nil,
			want: []*Instance{{}},
		},
		{
			args: []string{},
			want: []*Instance{
				{Args: []string{}},
			},
		},
		{
			args: []string{""},
			want: []*Instance{
				{Args: []string{""}},
			},
		},
		{
			args: []string{"--"},
			want: []*Instance{
				{Args: []string{}},
				{Args: []string{}},
			},
		},
		{
			args: []string{"--", "arg1", "--"},
			want: []*Instance{
				{Args: []string{}},
				{Args: []string{"arg1"}},
				{Args: []string{}},
			},
		},
		{
			args: []string{"arg1", "--"},
			want: []*Instance{
				{Args: []string{"arg1"}},
				{Args: []string{}},
			},
		},
		{
			args: []string{"arg1"},
			want: []*Instance{{Args: []string{"arg1"}}},
		},
		{
			args: []string{"-flag1"},
			want: []*Instance{{Args: []string{"-flag1"}}},
		},
		{
			args: []string{"--flag1"},
			want: []*Instance{{Args: []string{"--flag1"}}},
		},
		{
			args: []string{"arg1 -- arg2"},
			want: []*Instance{{Args: []string{"arg1 -- arg2"}}},
		},
		{
			args: []string{"arg1", "--", "arg2"},
			want: []*Instance{
				{Args: []string{"arg1"}},
				{Args: []string{"arg2"}},
			},
		},
		{
			args: []string{"arg1", "-flag1", "arg2", "--flag2"},
			want: []*Instance{{Args: []string{"arg1", "-flag1", "arg2", "--flag2"}}},
		},
		{
			args: []string{"arg1", "-flag1", "--", "arg2", "--flag2"},
			want: []*Instance{
				{Args: []string{"arg1", "-flag1"}},
				{Args: []string{"arg2", "--flag2"}},
			},
		},
		{
			args: []string{"arg1", "-flag1", "--", "arg2", "--", "arg3", "-flag3"},
			want: []*Instance{
				{Args: []string{"arg1", "-flag1"}},
				{Args: []string{"arg2"}},
				{Args: []string{"arg3", "-flag3"}},
			},
		},
	}

	for index, test := range tests {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			got := parseArgs(test.args)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestParseWatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		watch   string
		max     int
		want    []int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			watch:   "",
			max:     3,
			want:    []int{},
			wantErr: assert.NoError,
		},
		{
			watch:   "none",
			max:     3,
			want:    []int{},
			wantErr: assert.NoError,
		},
		{
			watch:   "all",
			max:     3,
			want:    []int{0, 1, 2},
			wantErr: assert.NoError,
		},
		{
			watch:   "  1  ",
			max:     3,
			want:    []int{1},
			wantErr: assert.NoError,
		},
		{
			watch:   "0, 2",
			max:     3,
			want:    []int{0, 2},
			wantErr: assert.NoError,
		},
		{
			watch:   ",2",
			max:     3,
			want:    []int{2},
			wantErr: assert.NoError,
		},
		{
			watch:   "2, ",
			max:     3,
			want:    []int{2},
			wantErr: assert.NoError,
		},
		{
			watch:   "0,,2",
			max:     3,
			want:    []int{0, 2},
			wantErr: assert.NoError,
		},
		{
			watch:   "-1",
			max:     3,
			want:    nil,
			wantErr: assert.Error,
		},
		{
			watch:   "3",
			max:     3,
			want:    nil,
			wantErr: assert.Error,
		},
		{
			watch:   "4",
			max:     3,
			want:    nil,
			wantErr: assert.Error,
		},
		{
			watch:   "a",
			max:     3,
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for index, test := range tests {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			got, err := parseInts(test.watch, test.max)
			assert.Equal(t, test.want, got)
			test.wantErr(t, err)
		})
	}
}
