package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	_ "pfe/docs"
	"pfe/internal/api"
	"pfe/internal/transfer"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Firego API
// @version 1.0
// @description 대용량 파일 실시간 전송 엔진
// @host localhost:54321
// @BasePath /
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

	if os.Getenv("ENV") != "production" {
		http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
		log.Println("Swagger UI: http://localhost:54321/swagger/index.html")
	}

	log.Println("API SERVER START... PORT: 54321")

	if err := http.ListenAndServe(":54321", nil); err != nil {
		log.Fatal(err)
	}
}
