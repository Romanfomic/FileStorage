package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/config"
	"backend/routes"
)

func main() {
	// Подключаем MongoDB
	config.ConnectDB()

	// Регистрируем маршруты
	r := routes.RegisterRoutes()

	// Запускаем сервер
	port := ":8080"
	fmt.Println("🚀 Сервер запущен на", port)
	log.Fatal(http.ListenAndServe(port, r))
}
