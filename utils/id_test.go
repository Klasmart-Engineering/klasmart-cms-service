package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewId(t *testing.T) {
	t.Log(NewID())
	t.Log(NewID())
	t.Log(len(NewID()))
	t.Log(NewID())
}

func TestUniqueIdList(t *testing.T) {
	r := SliceDeduplication([]string{"a", "a", "a", "a", "a", "a"})
	fmt.Println(r)
}

func TestNotValidUUID(t *testing.T) {
	assert.False(t, IsValidUUID(""))
}

func TestValidUUID(t *testing.T) {
	assert.True(t, IsValidUUID("d653f880-b920-435f-8f2d-98b5d4621eb8"))
}
