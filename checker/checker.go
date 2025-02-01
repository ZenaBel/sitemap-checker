package checker

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"sitemap-checker/config"
	"sitemap-checker/fetcher"
	"sitemap-checker/logger"
	"sitemap-checker/parser"
)

// PageResult містить результати перевірки сторінки
type PageResult struct {
	URL                  string            `json:"url"`
	StatusCode           int               `json:"status_code"`
	Redirects            []string          `json:"redirects"`
	CanonicalURL         string            `json:"canonical_url"`
	MetaTags             map[string]string `json:"meta_tags"`
	LoadTime             string            `json:"load_time"`
	IsBlockedByRobotsTxt bool              `json:"is_blocked_by_robots_txt"`
	ContentHash          string            `json:"content_hash"`
}

var (
	results      []PageResult // Зберігаємо результати перевірок
	resultsMutex sync.Mutex   // Для потокобезпечного доступу
)

// ProcessURLSet обробляє список URL-адрес
func ProcessURLSet(ctx context.Context, urlset *parser.URLSet, wg *sync.WaitGroup, sem chan struct{}, cfg *config.Config, contentHashes map[string]string) {
	defer wg.Done()

	for _, url := range urlset.URLs {
		select {
		case <-ctx.Done():
			logger.Error("обробка перервана через скасування контексту")
			return
		default:
			sem <- struct{}{}
			wg.Add(1)
			go func(url parser.URL) {
				defer func() { <-sem }()

				// Перевірка robots.txt
				isAllowed := CheckRobotsTxt(ctx, url.Loc)

				// Завантажуємо сторінку з вимірюванням часу
				resp, redirects, loadTime, err := fetcher.FetchPageWithTiming(ctx, url.Loc, cfg.MaxRedirects)
				if err != nil {
					logger.Error("помилка при завантаженні сторінки %s: %v", url.Loc, err)
					wg.Done()
					return
				}
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						logger.Error("помилка при закритті тіла відповіді: %v", err)
					}
				}(resp.Body)

				// Перевірка часу завантаження
				CheckPageLoadTime(url.Loc, loadTime, 2*time.Second) // Поріг: 2 секунди

				// Читання тіла сторінки
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					logger.Error("помилка при читанні тіла сторінки: %v", err)
					wg.Done()
					return
				}

				// Перевірка на дублі контенту
				contentHash := CheckContentDuplicates(string(body), contentHashes, url.Loc)

				// Збір даних про сторінку
				pageResult := PageResult{
					URL:                  url.Loc,
					StatusCode:           resp.StatusCode,
					Redirects:            redirects,
					CanonicalURL:         extractCanonicalURL(string(body)),
					MetaTags:             extractMetaTags(string(body)),
					LoadTime:             loadTime.String(),
					IsBlockedByRobotsTxt: !isAllowed,
					ContentHash:          contentHash,
				}

				// Зберігаємо результат
				resultsMutex.Lock()
				results = append(results, pageResult)
				resultsMutex.Unlock()

				wg.Done()
			}(url)
		}
	}
}

// extractCanonicalURL витягує канонічне посилання з HTML
func extractCanonicalURL(html string) string {
	if strings.Contains(html, `<link rel="canonical"`) {
		// Проста реалізація для прикладу
		start := strings.Index(html, `href="`) + 6
		end := strings.Index(html[start:], `"`) + start
		return html[start:end]
	}
	return ""
}

// extractMetaTags витягує мета-теги з HTML
func extractMetaTags(html string) map[string]string {
	metaTags := make(map[string]string)

	// Вилучення тегу <title>
	if titleStart := strings.Index(html, "<title>"); titleStart != -1 {
		titleEnd := strings.Index(html[titleStart:], "</title>")
		if titleEnd != -1 {
			metaTags["title"] = html[titleStart+7 : titleStart+titleEnd]
		}
	}

	// Вилучення мета-тегу description
	if descStart := strings.Index(html, `<meta name="description"`); descStart != -1 {
		contentStart := strings.Index(html[descStart:], `content="`) + descStart + 9
		contentEnd := strings.Index(html[contentStart:], `"`) + contentStart
		metaTags["description"] = html[contentStart:contentEnd]
	}

	return metaTags
}

// CheckRobotsTxt перевіряє, чи сторінка дозволена в robots.txt
func CheckRobotsTxt(ctx context.Context, pageURL string) bool {
	robotsTxt, err := fetcher.FetchRobotsTxt(ctx, pageURL)
	if err != nil {
		logger.Error("помилка при завантаженні robots.txt: %v", err)
		return true // Якщо robots.txt недоступний, вважаємо сторінку дозволеною
	}

	// Перевіряємо, чи сторінка дозволена
	if strings.Contains(string(robotsTxt), "Disallow: "+pageURL) {
		logger.Error("сторінка заблокована в robots.txt: %s", pageURL)
		return false
	}

	return true
}

// CheckPageLoadTime перевіряє час завантаження сторінки
func CheckPageLoadTime(pageURL string, loadTime time.Duration, threshold time.Duration) {
	if loadTime > threshold {
		logger.Error("сторінка завантажується повільно: %s (час: %v)", pageURL, loadTime)
	} else {
		logger.Info("сторінка завантажена швидко: %s (час: %v)", pageURL, loadTime)
	}
}

// CheckContentDuplicates перевіряє наявність дублів контенту
func CheckContentDuplicates(content string, contentHashes map[string]string, pageURL string) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	if existingURL, exists := contentHashes[hash]; exists {
		logger.Error("дубль контенту: %s та %s", pageURL, existingURL)
	} else {
		contentHashes[hash] = pageURL
	}
	return hash
}

// CheckStatusCode перевіряє статус-код сторінки
func CheckStatusCode(resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		logger.Error("неправильний статус-код: %d для URL: %s", resp.StatusCode, resp.Request.URL)
	}
}

// CheckRedirects перевіряє редіректи
func CheckRedirects(redirects []string, originalURL string) {
	if len(redirects) > 0 {
		logger.Error("редіректи для URL %s: %v", originalURL, redirects)
	} else {
		logger.Info("редіректи відсутні для URL: %s", originalURL)
	}
}

// CheckCanonicalLink перевіряє канонічне посилання
func CheckCanonicalLink(resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("помилка при читанні тіла відповіді: %v", err)
		return
	}

	// Шукаємо канонічне посилання
	if strings.Contains(string(body), `<link rel="canonical"`) {
		logger.Info("канонічне посилання знайдено для URL: %s", resp.Request.URL)
	} else {
		logger.Error("канонічне посилання відсутнє для URL: %s", resp.Request.URL)
	}
}

// CheckMetaTags перевіряє мета-теги
func CheckMetaTags(resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("помилка при читанні тіла відповіді: %v", err)
		return
	}

	// Перевірка наявності тегу <title>
	if !strings.Contains(string(body), "<title>") {
		logger.Error("тег <title> відсутній для URL: %s", resp.Request.URL)
	}

	// Перевірка наявності метатегу description
	if !strings.Contains(string(body), `<meta name="description"`) {
		logger.Error("мета-тег description відсутній для URL: %s", resp.Request.URL)
	}
}

// ProcessSitemapIndex обробляє вкладені файли sitemap
func ProcessSitemapIndex(ctx context.Context, sitemapIndex *parser.SitemapIndex, depth int, wg *sync.WaitGroup, sem chan struct{}, cfg *config.Config, contentHashes map[string]string) {
	defer wg.Done()

	if depth > cfg.MaxDepth {
		logger.Error("досягнуто максимальну глибину рекурсії: %d", depth)
		return
	}

	for _, sitemap := range sitemapIndex.Sitemaps {
		select {
		case <-ctx.Done():
			logger.Error("обробка перервана через скасування контексту")
			return
		default:
			sem <- struct{}{}
			wg.Add(1)
			go func(sitemap parser.SitemapURL) {
				defer func() { <-sem }()

				// Завантажуємо та обробляємо кожен файл sitemap
				data, err := fetcher.FetchSitemap(ctx, sitemap.Loc)
				if err != nil {
					logger.Error("помилка при завантаженні файлу sitemap %s: %v", sitemap.Loc, err)
					wg.Done()
					return
				}

				sitemapContent, err := parser.ParseSitemap(data)
				if err != nil {
					logger.Error("помилка при парсингу файлу sitemap %s: %v", sitemap.Loc, err)
					wg.Done()
					return
				}

				switch content := sitemapContent.(type) {
				case *parser.URLSet:
					wg.Add(1)
					ProcessURLSet(ctx, content, wg, sem, cfg, contentHashes)
				case *parser.SitemapIndex:
					wg.Add(1)
					ProcessSitemapIndex(ctx, content, depth+1, wg, sem, cfg, contentHashes)
				default:
					logger.Error("невідомий тип вмісту sitemap: %s", sitemap.Loc)
				}
				wg.Done()
			}(sitemap)
		}
	}
}

// SaveResultsToJSON зберігає результати у JSON-файл
func SaveResultsToJSON(filename string) error {
	resultsMutex.Lock()
	defer resultsMutex.Unlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("помилка при створенні файлу: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("помилка при записі JSON: %v", err)
	}

	return nil
}
