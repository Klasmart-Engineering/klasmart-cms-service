package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestMarshalConfig(t *testing.T){
	conf := new(Config)
	confStr, err := yaml.Marshal(conf)
	if err != nil{
		panic(err)
	}
	fmt.Println(string(confStr))
}
