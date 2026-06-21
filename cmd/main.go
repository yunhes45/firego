package main

import (
	"fmt"
	"log"
	"net/http"
	"pfe/internal/api"
	"pfe/internal/transfer"
)

func main() {
	fmt.Println("File Transfer Engine Start...")

	sm := transfer.NewSessionManager()
	stm := transfer.NewStreamManager()
	h := api.NewHandler(sm, stm)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	http.HandleFunc("/session", h.CreateSession)
	http.HandleFunc("/send/", h.Send)
	http.HandleFunc("/receive/", h.Receive)
	log.Println("API SERVER START... PORT: 54321")

	if err := http.ListenAndServe(":54321", nil); err != nil {
		log.Fatal(err)
	}
}
