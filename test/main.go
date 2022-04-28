package main

import (
	"net/http"

	"github.com/Jetereting/swaggo/test/pkg"
)

func main() {
	router := pkg.New()
	http.ListenAndServe("localhost:3000", router)
}
