package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hola mundo desde Go en el puerto 8082!")
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Servidor escuchando en http://localhost:8082")
    http.ListenAndServe(":8082", nil)
}