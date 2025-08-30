package config

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/ary/go-api/models"
	"github.com/joho/godotenv" // tambahkan ini

	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Load .env (hanya jika belum jalan di Docker / CI)
	if os.Getenv("RUNNING_IN_DOCKER") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️  .env file not found, using environment variables from system")
		}
	}

	// Gunakan variabel dari .env / system
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	fmt.Println("🔍 DSN:", dsn) // <-- Tambahkan ini

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	DB = db

	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderStatusUpdate{},
		&models.UserAccount{},
		&models.Notification{},
		&models.Info{},
	)

	if err != nil {
		log.Println("❌ Failed to auto-migrate:", err)
	}

	fmt.Println("📦 Connected to DB")
}

// contoh redis config
var RedisClient *redis.Client

func ConnectRedis() {
	// Default TLS config kosong (untuk lokal)
	var tlsConfig *tls.Config = nil

	// Kalau mau konek ke Redis yang butuh TLS (contoh Railway)
	if os.Getenv("REDIS_TLS") == "true" {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, // boleh diaktifkan kalau self-signed
		}
	}

	opt := &redis.Options{
		Addr:      fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Username:  os.Getenv("REDIS_USERNAME"),
		Password:  os.Getenv("REDIS_PASSWORD"),
		DB:        0,
		TLSConfig: tlsConfig,
	}

	client := redis.NewClient(opt)

	// Ping untuk tes koneksi
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Println("⚠️ Redis connection failed:", err)
		return // Jangan matikan server, biarkan jalan
	}

	RedisClient = client
	fmt.Println("🔌 Connected to Redis")
}
