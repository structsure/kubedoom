package main

import (
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {

	got := hash("The Quick Brown Fox Jumps Over The Lazy Dog")
	var want int32 = 1467600862

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
func TestAddString(t *testing.T) {
	tryAToI("", t.Log)
	tryAToI("0", t.Log)
}

func tryAToI(astr string, f func(args ...any)) {
	val, _ := strconv.Atoi(astr)
	f("%v == %v", astr, val)
}
