package utils

import "testing"

func TestNewId(t *testing.T) {
	t.Log(NewID())
	t.Log(NewID())
	t.Log(NewID())
	t.Log(NewID())
}
