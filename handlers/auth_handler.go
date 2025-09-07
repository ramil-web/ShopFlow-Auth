package handlers

import (
	"auth/models"
	"auth/services"
	"crypto/rand"
	_ "encoding/json"
	"log"
	_ "log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// RegisterHandler
// @Summary User registration
// @Description Creates a new user account and returns the created user object
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "Registration data"
// @Success 201 {object} models.User "User successfully registered"
// @Failure 400 {object} map[string]string "Invalid input data"
// @Failure 409 {object} map[string]string "User already exists"
// @Router /api/auth/register [post]
func RegisterHandler(db *gorm.DB, publisher *services.EventPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Генерация случайного пароля
		pass, err := generateRandomPassword(8)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate password"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		user := models.User{
			Login:    req.Email,
			Email:    req.Email,
			Password: string(hash),
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		go func() {
			if err := publisher.PublishUserRegistered(user.ID, user.Email, user.Login, pass); err != nil {
				log.Println("[ERROR] failed to publish user.registered event:", err)
			}
		}()

		c.JSON(http.StatusCreated, gin.H{
			"id":    user.ID,
			"login": user.Login,
			"email": user.Email,
		})
	}
}

// Генерация случайного пароля
func generateRandomPassword(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		password[i] = chars[num.Int64()]
	}
	return string(password), nil
}

// LoginHandler @Summary Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "User login"
// @Success 200 {object} TokenResponse
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.Where("login = ?", req.Login).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"login":   user.Login,
			"exp":     time.Now().Add(720 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, TokenResponse{Token: tokenString})
	}
}
