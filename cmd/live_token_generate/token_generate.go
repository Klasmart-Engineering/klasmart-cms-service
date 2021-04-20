package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"io/ioutil"
	"os"
)

func main() {
	jsonFilePath := ""
	privateKeyFilePath := ""
	fmt.Println("Please enter private key file path: ")
	fmt.Scanln(&privateKeyFilePath)
	b, err := ReadAll(privateKeyFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	privateKey := string(b)

	fmt.Println("Please enter json file path: ")
	fmt.Scanln(&jsonFilePath)
	b2, err := ReadAll(jsonFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonStr := string(b2)
	token, err := GenerateByJsonString(jsonStr, privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(token)
}

func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func GenerateByJsonString(jsonStr string, privateKey string) (string, error) {
	data := new(entity.LiveTokenClaims)
	err := json.Unmarshal([]byte(jsonStr), data)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	pb, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		fmt.Println("ParseRSAPublicKeyFromPEM:", err.Error())
		return "", err
	}
	token, err := utils.CreateJWT(context.Background(), data, pb)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return token, nil
}
