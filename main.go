package main

import (
	"fmt"
	"net/http"
)

func main(){
	mux := http.NewServeMux()

	server := http.Server{
		Handler: mux,
		Addr: ":8080",
	}


	mux.Handle("/", http.FileServer(http.FileSystem(http.Dir("."))))


	err := server.ListenAndServe()

	if err != nil{
		fmt.Println("Couldn't run server!")
		return
	}


	return
}
