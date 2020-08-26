package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var server *Server

func TestMain(m *testing.M) {
	server = NewServer()
	code := m.Run()
	os.Exit(code)
}

const url = ""
const prefix = "/v1"

func DoHttp(method string, url string, body string) *http.Response {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Add("Authorization", "")
	server.ServeHTTP(w, req)
	res := w.Result()
	return res
}

func TestApprove(t *testing.T) {
	res := DoHttp(http.MethodPut, prefix+"/contents_review/1/approve", "")
	fmt.Println(res)
}

func TestGetTimeLocation(t *testing.T) {
	//fmt.Println(time.LoadLocation("America/Los_Angeles"))
	loc := time.Local
	time.ParseInLocation("", "", loc)
	//fmt.Println(time.Local)
}
