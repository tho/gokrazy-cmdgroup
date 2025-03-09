package cmdgroup_test

import (
	"os/exec"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmdgroup "github.com/tho/gokrazy-cmdgroup"
)

func TestNew(t *testing.T) {
	t.Parallel()

	name := "echo"
	path, err := exec.LookPath(name)
	require.NoError(t, err)

	tests := []struct {
		name          string
		options       []cmdgroup.Option
		wantInstances []*cmdgroup.Instance
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name:    "",
			options: nil,
			wantErr: assert.Error,
		},
		{
			name:    name,
			options: []cmdgroup.Option{cmdgroup.WithWatch("a")},
			wantErr: assert.Error,
		},
		{
			name:          name,
			options:       nil,
			wantInstances: []*cmdgroup.Instance{{Name: path}},
			wantErr:       assert.NoError,
		},
		{
			name: name,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"arg1", "-flag1", "--", "arg2", "-flag2"}),
				cmdgroup.WithWatch("none"),
			},
			wantInstances: []*cmdgroup.Instance{
				{Name: path, Args: []string{"arg1", "-flag1"}, Watch: false},
				{Name: path, Args: []string{"arg2", "-flag2"}, Watch: false},
			},
			wantErr: assert.NoError,
		},
		{
			name: name,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"arg1", "-flag1", "--", "arg2", "-flag2"}),
				cmdgroup.WithWatch("all"),
			},
			wantInstances: []*cmdgroup.Instance{
				{Name: path, Args: []string{"arg1", "-flag1"}, Watch: true},
				{Name: path, Args: []string{"arg2", "-flag2"}, Watch: true},
			},
			wantErr: assert.NoError,
		},
		{
			name: name,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"arg1", "-flag1", "--", "arg2", "-flag2"}),
				cmdgroup.WithWatch("1"),
			},
			wantInstances: []*cmdgroup.Instance{
				{Name: path, Args: []string{"arg1", "-flag1"}, Watch: false},
				{Name: path, Args: []string{"arg2", "-flag2"}, Watch: true},
			},
			wantErr: assert.NoError,
		},
	}

	for index, test := range tests {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			group, err := cmdgroup.New(test.name, test.options...)
			test.wantErr(t, err)
			if group != nil {
				assert.EqualExportedValues(t, test.wantInstances, group.Instances)
			}
		})
	}
}
