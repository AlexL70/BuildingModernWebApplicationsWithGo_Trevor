package main

import "testing"

func TestRun(t *testing.T) {
	err := run()
	if err != nil {
		t.Errorf("Failed run: %q", err)
	}
}
