package utils

import (
	"context"
	"testing"
)

func TestURLSig(t *testing.T) {
	token, err := GenerateH5pJWT(context.Background(), "clone", "5fa3f41cb65ebc00122a74a6")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(token)
}
