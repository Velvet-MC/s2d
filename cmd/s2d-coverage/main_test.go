package main

import "testing"

func TestParseArgsDefaultsToJavaCoverage(t *testing.T) {
	cfg, err := parseArgs(nil)
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if !cfg.java {
		t.Fatal("java coverage should be enabled by default")
	}
}
