package log

import (
	"testing"
)

func TestTimeString(t *testing.T){
	log := StdLog{}

	log.Error("Hello")
	log.Errorf("Hello%d", 100)
}
