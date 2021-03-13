package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

var lineMap map[string]bool

func handler(line string) {
	lineMap[line] = true
}

func TestFilterData(t *testing.T) {
	lineMap = make(map[string]bool)

	f, err := os.Open("./data")
	if err != nil {
		t.Error(err)
		return
	}

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line)
		if err != nil {
			if err == io.EOF {
				break
			}

			t.Error(err)
			return
		}
	}

	for k := range lineMap {
		fmt.Println(k)
		// if strings.HasPrefix(k, "can't find program") {
		// 	fmt.Println(k)
		// }
		// if strings.HasPrefix(k, "can't find subjects") {
		// 	fmt.Println(k)
		// }

	}
}
