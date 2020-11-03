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

func TestMockClassService_BatchGet(t *testing.T) {
	classes, err := GetClassServiceProvider().BatchGet(context.Background(), []string{
		"f3d3cdf5-9ca8-44cf-a604-482e5d183049",
		"3b8074d5-893c-41c7-942f-d2115cc8bc32",
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := range classes {
		fmt.Println(*(classes[i]))
	}
}

func TestMockClassService_GetStudents(t *testing.T) {
	students, err := GetClassServiceProvider().GetStudents(context.Background(), "f3d3cdf5-9ca8-44cf-a604-482e5d183049")
	if err != nil {
		t.Fatal(err)
	}
	for i := range students {
		fmt.Println(*(students[i]))
	}
}
