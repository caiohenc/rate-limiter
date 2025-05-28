package main

import (
	"log"
	"net/http"
	"os"
	"rate-limiter/limiter"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar configurações
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found - using default values")
	}

	// Configurar rate limiter
	rateIP, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_IP"))
	rateToken, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_TOKEN"))
	blockTime, _ := strconv.Atoi(os.Getenv("BLOCK_TIME"))
	port := os.Getenv("SERVER_PORT")

	if port == "" {
		port = "8080"
	}

	// Inicializar Redis
	redisStorage := limiter.NewRedisStorage(
		os.Getenv("REDIS_ADDR"),
		os.Getenv("REDIS_PASSWORD"),
		os.Getenv("REDIS_DB"),
	)

	// Criar rate limiter
	limiter := limiter.NewLimiter(
		redisStorage,
		rateIP,
		rateToken,
		time.Duration(blockTime)*time.Second,
	)

	// Configurar servidor
	router := gin.Default()
	router.Use(limiter.Middleware())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Rate Limited API",
			"status":  "OK",
		})
	})

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		if err := redisStorage.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "DOWN",
				"message": "Redis down",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "UP",
			"message": "Service up",
		})
	})

	log.Printf("Server running on port %s", port)
	log.Fatal(router.Run(":" + port))
}
