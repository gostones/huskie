package server

import (
	"fmt"
	"net/http"
)

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong!")
}
