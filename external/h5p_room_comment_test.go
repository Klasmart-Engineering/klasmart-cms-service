package external

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestH5PRoomCommentService_BatchGet(t *testing.T) {
	comments, err := GetH5PRoomCommentServiceProvider().BatchGet(context.TODO(),
		testOperator,
		[]string{"60a1d40a03b03c3acdb4f946"})
	if err != nil {
		t.Errorf("GetH5PRoomCommentServiceProvider().BatchGet() error = %v", err)
		return
	}

	if len(comments) == 0 {
		t.Error("GetH5PRoomCommentServiceProvider().BatchGet() get empty slice")
		return
	}

	for _, comment := range comments {
		if len(comment) == 0 {
			t.Error("GetH5PRoomCommentServiceProvider().BatchGet() get null")
			return
		}
	}
}

func TestH5PRoomCommentService_Add(t *testing.T) {
	comment, err := GetH5PRoomCommentServiceProvider().Add(context.TODO(),
		testOperator,
		&H5PAddRoomCommentRequest{
			RoomID:    "60a1d40a03b03c3acdb4f946",
			StudentID: "6ef232ce-5c37-4550-a8ca-8d27da5133f8",
			Comment:   fmt.Sprintf("Comment-%s", time.Now().Format("20060102150405")),
		})
	if err != nil {
		t.Errorf("GetH5PRoomCommentServiceProvider().Add() error = %v", err)
		return
	}

	if comment == nil {
		t.Error("GetH5PRoomCommentServiceProvider().Add() get empty result")
		return
	}
}

func TestH5PRoomCommentService_BatchAdd(t *testing.T) {
	requests := []*H5PAddRoomCommentRequest{
		{
			RoomID:    "60a1d40a03b03c3acdb4f946",
			StudentID: "6ef232ce-5c37-4550-a8ca-8d27da5133f8",
			Comment:   fmt.Sprintf("CommentA-%s", time.Now().Format("20060102150405")),
		},
		{
			RoomID:    "60a1d40a03b03c3acdb4f946",
			StudentID: "6ef232ce-5c37-4550-a8ca-8d27da5133f8",
			Comment:   fmt.Sprintf("CommentB-%s n\n", time.Now().Format("20060102150405")),
		},
		{
			RoomID:    "60a1d40a03b03c3acdb4f946",
			StudentID: "6ef232ce-5c37-4550-a8ca-8d27da5133f8",
			Comment: `some
			text
			in
			multiple
			lines`,
		},
	}
	commentResults, err := GetH5PRoomCommentServiceProvider().BatchAdd(context.TODO(), testOperator, requests)
	if err != nil {
		t.Errorf("GetH5PRoomCommentServiceProvider().BatchAdd() error = %v", err)
		return
	}

	if len(commentResults) != len(requests) {
		t.Errorf("GetH5PRoomCommentServiceProvider().BatchAdd() get invalid result, want %d, got %d", len(requests), len(commentResults))
		return

	}
}
