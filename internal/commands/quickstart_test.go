package commands

import "testing"

func TestQuickstart(t *testing.T) {
	err := Quickstart([]string{})
	if err != nil {
		t.Fatalf("Quickstart() error = %v", err)
	}
}
