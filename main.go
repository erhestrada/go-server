package main

import "net/http"

func main() {
	serveMux := http.NewServeMux()
	x := http.Server{Addr: ":8080",
		Handler: serveMux}
	x.ListenAndServe()
}
