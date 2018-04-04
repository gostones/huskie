package server

import (
	"net/http"
	"fmt"
)

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong!")
}