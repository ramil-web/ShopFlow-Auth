package main

import (
	"auth/proto/authpb"
	"auth/routes"
	"auth/server"
	"auth/services"
	"fmt"
	"log"
	"net"
	"os"

	_ "auth/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Auth API
// @version 1.0
// @description API приложения Auth с JWT авторизацией
// @host localhost:8080
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// загрузка env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Настройка подключения к базе
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_DATABASE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	publisher := services.NewEventPublisher(conn)

	// Gin REST
	r := gin.Default()
	routes.SetupRoutes(r, db, publisher) // теперь db != nil

	// Swagger
	swaggerURL := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL, ginSwagger.PersistAuthorization(true)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запуск gRPC сервера в отдельной горутине
	go func() {
		grpcLis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen gRPC: %v", err)
		}

		grpcServer := grpc.NewServer()
		authpb.RegisterAuthServiceServer(grpcServer, &server.AuthService{})

		log.Println("gRPC Auth server running on :50051")
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Запуск REST
	log.Println("Auth REST service running on port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
