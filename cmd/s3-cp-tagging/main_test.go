package main

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// mockRunner implements a mock execution function for testing
type mockRunner struct {
	executed bool
	err      error
}

func (m *mockRunner) run(cmd *cobra.Command, args []string) error {
	m.executed = true
	return m.err
}

func TestMain(m *testing.M) {
	// Save original os.Exit
	originalExit := osExit
	defer func() { osExit = originalExit }()

	// Mock os.Exit
	osExit = func(code int) {
		// Do nothing to prevent test from exiting
	}

	// Run tests
	os.Exit(m.Run())
}

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "missing destination",
			args:    []string{"./source"},
			wantErr: true,
		},
		{
			name:    "missing tagging",
			args:    []string{"./source", "s3://bucket"},
			wantErr: true,
		},
		{
			name:    "valid args with tagging",
			args:    []string{"./source", "s3://bucket", "--tagging", "TagSet=[{Key=version,Value=v1.0.0}]"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRunner{}
			cmd := newRootCmd()
			// Replace command execution function with mock
			if !tt.wantErr {
				cmd.RunE = runner.run
			}
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check execution only for success cases
			if !tt.wantErr && !runner.executed {
				t.Error("Command should have been executed")
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("Version flag should not return error: %v", err)
	}
} 