package routes

import (
	"auth/handlers"
	"auth/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, publisher *services.EventPublisher) {
	api := r.Group("/api")
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.RegisterHandler(db, publisher))
		auth.POST("/login", handlers.LoginHandler(db)) // добавлен роут логина
	}
}
