package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/config"
	"backend/routes"
)

func main() {
	config.ConnectDB()

	r := routes.RegisterRoutes()

	port := ":8080"
	fmt.Println("🚀 Сервер запущен на", port)
	log.Fatal(http.ListenAndServe(port, r))
}
