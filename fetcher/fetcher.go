package fetcher

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sitemap-checker/logger"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
)

// InitRedis ініціалізує Redis клієнт
func InitRedis(redisURL string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // Пароль, якщо потрібно
		DB:       0,  // Номер бази даних
	})

	// Перевіряємо підключення до Redis
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Info("Redis недоступний, використовується тимчасовий файл для кешування")
		redisClient = nil
	}
}

// FetchSitemap завантажує sitemap.xml з вказаного URL
func FetchSitemap(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("помилка при створенні запиту: %v", err)
	}

	// Виконання запиту
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("помилка при завантаженні sitemap: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("помилка при закритті тіла відповіді: %v", err)
		}
	}(resp.Body)

	// Перевірка статус-коду
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неправильний статус код: %d", resp.StatusCode)
	}

	// Читання тіла відповіді
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("помилка при читанні тіла відповіді: %v", err)
	}

	return data, nil
}

// FetchPage завантажує сторінку з вказаного URL з підтримкою редіректів
func FetchPage(ctx context.Context, url string, maxRedirects int) (*http.Response, []string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("досягнуто максимальну кількість редіректів: %d", maxRedirects)
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("помилка при створенні запиту: %v", err)
	}

	// Виконання запиту
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("помилка при завантаженні сторінки: %v", err)
	}

	// Отримання ланцюжка редіректів
	redirects := make([]string, 0)
	if resp.Request != nil && resp.Request.Response != nil {
		for _, r := range resp.Request.Response.Header["Location"] {
			redirects = append(redirects, r)
		}
	}

	return resp, redirects, nil
}

// FetchRobotsTxt завантажує robots.txt з кешу або з мережі
func FetchRobotsTxt(ctx context.Context, pageURL string) ([]byte, error) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("помилка при парсингу URL: %v", err)
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	// Якщо Redis доступний, використовуємо його
	if redisClient != nil {
		cachedRobotsTxt, err := redisClient.Get(ctx, robotsURL).Bytes()
		if err == nil {
			return cachedRobotsTxt, nil
		}
	}

	// Якщо Redis недоступний, використовуємо тимчасовий файл
	tempFile := fmt.Sprintf("/tmp/robots_%x.txt", sha256.Sum256([]byte(robotsURL)))
	if data, err := os.ReadFile(tempFile); err == nil {
		return data, nil
	}

	// Якщо немає в кеші, завантажуємо з мережі
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("помилка при створенні запиту: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("помилка при завантаженні robots.txt: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("помилка при закритті тіла відповіді: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неправильний статус код: %d", resp.StatusCode)
	}

	robotsTxt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("помилка при читанні тіла відповіді: %v", err)
	}

	// Зберігаємо в кеш або тимчасовий файл
	if redisClient != nil {
		err = redisClient.Set(ctx, robotsURL, robotsTxt, 24*time.Hour).Err()
		if err != nil {
			logger.Error("помилка при збереженні robots.txt в кеш: %v", err)
		}
	} else {
		err = os.WriteFile(tempFile, robotsTxt, 0644)
		if err != nil {
			logger.Error("помилка при збереженні robots.txt у тимчасовий файл: %v", err)
		}
	}

	return robotsTxt, nil
}

// CleanupTempFiles видаляє тимчасові файли robots.txt
func CleanupTempFiles() {
	files, err := os.ReadDir("/tmp")
	if err != nil {
		logger.Error("помилка при читанні тимчасових файлів: %v", err)
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "robots_") {
			err := os.Remove("/tmp/" + file.Name())
			if err != nil {
				logger.Error("помилка при видаленні тимчасового файлу: %v", err)
			} else {
				logger.Info("тимчасовий файл видалено: %s", file.Name())
			}
		}
	}
}

// FetchPageWithTiming завантажує сторінку та вимірює час завантаження
func FetchPageWithTiming(ctx context.Context, pageURL string, maxRedirects int) (*http.Response, []string, time.Duration, error) {
	start := time.Now()

	resp, redirects, err := FetchPage(ctx, pageURL, maxRedirects)
	if err != nil {
		return nil, nil, 0, err
	}

	elapsed := time.Since(start)
	return resp, redirects, elapsed, nil
}
