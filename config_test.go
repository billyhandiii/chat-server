package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	expectations := []struct {
		filename    string
		shouldError bool
	}{
		{"invalid_ip.json", true},
		{"invalid_log.json", true},
		{"invalid_port.json", true},
		{"ip_log.json", false},
		{"ip_port.json", false},
		{"ip_port_log.json", false},
		{"nonexistent_log_file_path.json", false},
		{"nothing.json", false},
		{"only_ip.json", false},
		{"only_log_file_path.json", false},
		{"only_port.json", false},
		{"port_log.json", false},
	}

	for _, ex := range expectations {
		testName := ex.filename
		if ex.shouldError {
			testName += " expecting error"
		} else {
			testName += " not expecting error"
		}
		t.Run(testName, func(t *testing.T) {
			_, err := readConfigFile("test_data/configs/" + ex.filename)
			if ex.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
