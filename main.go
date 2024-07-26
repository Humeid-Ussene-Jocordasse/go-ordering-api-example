package main

import (
	"fmt"
	"net/http"
)

func main() {

	// router := chi.NewRouter()

	// router.Get("/hello", basicHandler)

	server := &http.Server{
		Addr:    ":3000",
		Handler: http.HandlerFunc(basicHandler),
	}

	err := server.ListenAndServe()

	if err != nil {
		fmt.Println("failed to listen to server", err)
	}
}

func basicHandler(w http.ResponseWriter, r *http.Request) {

	// Handle Get
	if r.URL.Path == "/foo" {
		// handle get foo
		w.Write([]byte("Foo Logic Implementation"))
		return
	}

	if r.Method == http.MethodPost {
		// Handle POST

	}

	w.Write([]byte("Hello, World!"))
}
