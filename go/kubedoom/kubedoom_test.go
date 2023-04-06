package main

import "testing"

func TestHash(t *testing.T) {

	got := hash("The Quick Brown Fox Jumps Over The Lazy Dog")
	var want int32 = 1467600862

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
