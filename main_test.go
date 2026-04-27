// main_test.go
package main

import "testing"

func TestAdd(t *testing.T) {
	result := Add(1, 2)
	expected := 3

	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}
}
