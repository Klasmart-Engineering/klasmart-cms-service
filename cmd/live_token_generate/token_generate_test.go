package main

import (
	"fmt"
	"testing"
)

func TestGenerateByJsonString(t *testing.T) {
	var keyPath = "D:\\workbench\\auth_private_key.pem"
	b, err := ReadAll(keyPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	key := string(b)

	var jsonPath = "D:\\workbench\\1.json"
	b2, err := ReadAll(jsonPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonStr := string(b2)

	token, err := GenerateByJsonString(jsonStr, key)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(token)
}
