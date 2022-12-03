package main

import (
	"os"
	"testing"
)

func TestParseArg(t *testing.T) {
	os.Args = []string{"um"}
	prg, err := ParseArg()
	if err == nil {
		t.Errorf("expected error but got nil")
	}

	os.Args = []string{"um", "sandmark.um"}
	prg, err = ParseArg()
	if prg != "sandmark.um" {
		t.Errorf("expected 'sandmark.um' but got '%s'", prg)
	}

	if err != nil {
		t.Errorf("expected no error but got '%s'", err)
	}
}
