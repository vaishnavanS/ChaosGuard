package cmd

import (
	"bytes"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs(args)

	err := RootCmd.Execute()
	return buf.String(), err
}

func TestCLICommands(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version command",
			args:    []string{"version"},
			wantErr: false,
		},
		{
			name:    "doctor command",
			args:    []string{"doctor"},
			wantErr: false,
		},
		{
			name:    "init command",
			args:    []string{"init"},
			wantErr: false,
		},
		{
			name:    "config command",
			args:    []string{"config"},
			wantErr: false,
		},
		{
			name:    "status command",
			args:    []string{"status"},
			wantErr: false,
		},
		{
			name:    "stop command",
			args:    []string{"stop"},
			wantErr: false,
		},
		{
			name:    "dashboard command",
			args:    []string{"dashboard", "--open=false"},
			wantErr: false,
		},
		{
			name:    "report command",
			args:    []string{"report", "-f", "json"},
			wantErr: false,
		},
		{
			name:    "attack command missing target and type",
			args:    []string{"attack"},
			wantErr: true,
		},
		{
			name:    "attack command success simulation",
			args:    []string{"attack", "-t", "web-server", "-a", "pause"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCommand(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeCommand(%v) returned error %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}
