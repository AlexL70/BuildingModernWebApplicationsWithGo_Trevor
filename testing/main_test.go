package main

import "testing"

var tests = []struct {
	name     string
	divident float32
	divisor  float32
	expected float32
	isErr    bool
}{
	{"valid data", 100.0, 10.0, 10.0, false},
	{"invalid data", 100.0, 0.0, 0.0, true},
	{"expect-5", 100.0, 20.0, 5.0, false},
}

func Test_division(t *testing.T) {
	for _, tt := range tests {
		got, err := divide(tt.divident, tt.divisor)
		if tt.isErr && err == nil {
			t.Errorf("%s: expected an error, but did not get one", tt.name)
		}
		if !tt.isErr && err != nil {
			t.Errorf("%s: unexpected error: %q", tt.name, err)
		}
		if got != tt.expected {
			t.Errorf("%s: wrong result; expected %f, but got %f", tt.name, tt.expected, got)
		}
	}
}

func Test_divide(t *testing.T) {
	_, err := divide(10.0, 1.0)
	if err != nil {
		t.Error("Unexpected error")
	}
}

func Test_bad_divide(t *testing.T) {
	_, err := divide(10.0, 0.0)
	if err == nil {
		t.Error("Division by zero error was missed")
	}
}
