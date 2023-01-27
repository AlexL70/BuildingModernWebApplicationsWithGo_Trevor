package main

import (
	"net/http"
	"os"
	"testing"
)

type myHandler struct {
}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
