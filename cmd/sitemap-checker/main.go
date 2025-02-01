package main

import (
	"context"
	"sitemap-checker/checker"
	"sitemap-checker/config"
	"sitemap-checker/fetcher"
	"sitemap-checker/logger"
	"sitemap-checker/parser"
	"sync"
)

func main() {
	// Ініціалізація логера
	logger.Init("errors.log")
	defer logger.Close() // Закриваємо файл логів після завершення

	// Завантаження конфігурації
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Помилка при завантаженні конфігурації: %v", err)
		return
	}

	// Ініціалізація Redis
	fetcher.InitRedis(cfg.RedisURL)

	// Контекст з таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Завантаження sitemap.xml
	data, err := fetcher.FetchSitemap(ctx, cfg.SitemapURL)
	if err != nil {
		logger.Error("Помилка при завантаженні sitemap: %v", err)
		return
	}

	// Парсинг sitemap.xml
	sitemapContent, err := parser.ParseSitemap(data)
	if err != nil {
		logger.Error("Помилка при парсингу sitemap: %v", err)
		return
	}

	// Канал для обмеження кількості паралельних goroutines
	sem := make(chan struct{}, cfg.MaxGoroutines)
	var wg sync.WaitGroup

	// Мапа для зберігання хешів контенту
	contentHashes := make(map[string]string)

	// Обробка вмісту sitemap
	switch content := sitemapContent.(type) {
	case *parser.URLSet:
		wg.Add(1)
		checker.ProcessURLSet(ctx, content, &wg, sem, cfg, contentHashes)
	case *parser.SitemapIndex:
		wg.Add(1)
		checker.ProcessSitemapIndex(ctx, content, 1, &wg, sem, cfg, contentHashes)
	default:
		logger.Error("Невідомий тип вмісту sitemap")
	}

	wg.Wait()
	logger.Info("Перевірка завершена.")
}
