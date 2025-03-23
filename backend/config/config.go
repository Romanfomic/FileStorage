package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database // Mongo
var PostgresDB *sql.DB // PostgreSQL

func ConnectDB() {
	connectMongo()
	connectPostgres()
}

func connectMongo() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017/filestorage"
	}

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("MongoDB не отвечает:", err)
	}

	fmt.Println("✅ Подключение к MongoDB успешно!")

	DB = client.Database("filestorage")
}

func connectPostgres() {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "filestorage"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к PostgreSQL:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("PostgreSQL не отвечает:", err)
	}

	fmt.Println("✅ Подключение к PostgreSQL успешно!")

	PostgresDB = db
}
