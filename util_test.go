package main

import (
	"testing"
)

func testEscaperValue(t testing.T) {
	if escaperValue("") != "" {
		t.Error()
	}
}