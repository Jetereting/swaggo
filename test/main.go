package main

import (
	"net/http"

	"swaggo/test/pkg"
)

func main() {
	router := pkg.New()
	http.ListenAndServe("localhost:3000", router)
}
