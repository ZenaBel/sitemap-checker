package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	SitemapURL    string        // URL до sitemap.xml
	Timeout       time.Duration // Таймаут для HTTP-запитів
	MaxGoroutines int           // Максимальна кількість паралельних goroutines
	MaxDepth      int           // Максимальна глибина рекурсії для sitemapindex
	MaxRedirects  int           // Максимальна кількість редіректів
	RedisURL      string        // URL для підключення до Redis
}

func Load() (*Config, error) {
	// Завантажуємо змінні з .env файлу
	err := godotenv.Load()
	if err != nil {
		log.Println("Не вдалося завантажити .env файл, використовуються значення за замовчуванням")
	}

	// Таймаут
	timeoutStr := os.Getenv("TIMEOUT")
	timeout := 30 * time.Second // Значення за замовчуванням
	if timeoutStr != "" {
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("невірний формат таймауту: %v", err)
		}
	}

	// Максимальна кількість goroutines
	maxGoroutinesStr := os.Getenv("MAX_GOROUTINES")
	maxGoroutines := 10 // Значення за замовчуванням
	if maxGoroutinesStr != "" {
		maxGoroutines, err = strconv.Atoi(maxGoroutinesStr)
		if err != nil {
			return nil, fmt.Errorf("невірний формат MAX_GOROUTINES: %v", err)
		}
	}

	// Максимальна глибина рекурсії
	maxDepthStr := os.Getenv("MAX_DEPTH")
	maxDepth := 10 // Значення за замовчуванням
	if maxDepthStr != "" {
		maxDepth, err = strconv.Atoi(maxDepthStr)
		if err != nil {
			return nil, fmt.Errorf("невірний формат MAX_DEPTH: %v", err)
		}
	}

	// Максимальна кількість редіректів
	maxRedirectsStr := os.Getenv("MAX_REDIRECTS")
	maxRedirects := 5 // Значення за замовчуванням
	if maxRedirectsStr != "" {
		maxRedirects, err = strconv.Atoi(maxRedirectsStr)
		if err != nil {
			return nil, fmt.Errorf("невірний формат MAX_REDIRECTS: %v", err)
		}
	}

	// URL до Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis:6379" // Значення за замовчуванням для Docker
	}

	return &Config{
		SitemapURL:    os.Getenv("SITEMAP_URL"),
		Timeout:       timeout,
		MaxGoroutines: maxGoroutines,
		MaxDepth:      maxDepth,
		MaxRedirects:  maxRedirects,
		RedisURL:      redisURL,
	}, nil
}
