package external

import (
	"context"
	"fmt"
	"testing"
)

func TestMockOrganizationService_BatchGet(t *testing.T) {
	GetOrganizationServiceProvider().BatchGet(context.Background(), []string{
		"3f135b91-a616-4c80-914a-e4463104dbac",
		"3f135b91-a616-4c80-914a-e4463104dbad",
	})
}

func TestEmptySlice(t *testing.T) {
	sli := make([]string, 0)
	var sli2 []string
	if sli == nil {
		fmt.Println("sli nil")
	}

	if sli2 == nil {
		fmt.Println("sli2 nil")
	}
}
