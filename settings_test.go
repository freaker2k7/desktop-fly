package main

import (
	"os"
	"testing"
)

func TestGetenvHelpers(t *testing.T) {
	os.Setenv("FOO_TEST_INT", "33")
	defer os.Unsetenv("FOO_TEST_INT")
	v := getenv("NON_EXISTENT_KEY", "defval")
	if v != "defval" {
		t.Fatalf("expected default, got %s", v)
	}
	vi := getenvInt("FOO_TEST_INT", 5)
	if vi != 33 {
		t.Fatalf("expected 33, got %d", vi)
	}
}
