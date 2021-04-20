package main

import (
	"fmt"
	"net/url"
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

func TestUrl(t *testing.T) {
	escape := url.QueryEscape("https://res-kl2-test.kidsloop.net/assets/6062d0439a5d6267783deb27.ppt?Expires=1617175908&Signature=EpErl7Z~yJ840fwrxGgXetv0tZIUHpbQHBuYxKS2YmJ0RLHdqBiteAxcIN-Mu7OuBHkPWH6xFVuACpZ9I3~5NXC7Dr~0-UAi~4yvJ5gU09995XWqBMySvRwI8btlyqMcMsQ9oXDVc1UlW~Rvhlx7O7Uom7hG801UM-qJnLoMtu6TfoqYIBNJl6E21uJRoNd5dpO~NA-MOdA~Ds08qoR0pm2PM-oZ8sSabKzjmcyX8w-ND6k7~Qvsdimyqq3-xfCKZ05XBdJL1SEBWP~8nWmmkON1li7WuWgWVfJl1Qu2PuIjCJxL9cJhLwEtuF1NXCvxthuOylPgTudlChrx0LT8sg__&Key-Pair-Id=K3PUGKGK3R1NHM")
	t.Log(escape)
}
