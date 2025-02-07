package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewReceiptStore()
	handler := NewReceiptServer(store)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
