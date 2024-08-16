package main

import (
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/storage"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из файла .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	storage, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	r := mux.NewRouter()
	handler := &handlers.Handler{Storage: storage}
	authHandler := &handlers.AuthHandler{Storage: storage}

	// Роутинг
	r.HandleFunc("/", handler.HelloHandler).Methods("GET", "POST")
	r.HandleFunc("/create", handler.CreateShortLinkHandler).Methods("POST")
	r.HandleFunc("/{shortName:[a-zA-Z0-9]+}", handler.RedirectHandler).Methods("GET")
	r.HandleFunc("/register", authHandler.RegisterHandler).Methods("POST")
	r.HandleFunc("/admin/delete", authHandler.AdminDeleteUserHandler).Methods("POST")
	r.HandleFunc("/admin/users", authHandler.AdminGetUsersHandler).Methods("GET")

	// Получение порта из переменных окружения
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080" // Значение по умолчанию, если переменная окружения не установлена
	}

	// Запуск сервера
	log.Printf("url-shoterer is running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
