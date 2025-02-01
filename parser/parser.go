package parser

import (
	"encoding/xml"
	"fmt"
)

// URLSet представляє <urlset> у sitemap.xml
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

// SitemapIndex представляє <sitemapindex> у sitemap.xml
type SitemapIndex struct {
	XMLName  xml.Name     `xml:"sitemapindex"`
	Sitemaps []SitemapURL `xml:"sitemap"`
}

// URL представляє окремий URL у <urlset>
type URL struct {
	Loc        string  `xml:"loc"`        // URL сторінки
	LastMod    string  `xml:"lastmod"`    // Дата останньої зміни
	ChangeFreq string  `xml:"changefreq"` // Частота оновлення
	Priority   float32 `xml:"priority"`   // Пріоритет сторінки
}

// SitemapURL представляє окремий файл sitemap у <sitemapindex>
type SitemapURL struct {
	Loc     string `xml:"loc"`     // URL файлу sitemap
	LastMod string `xml:"lastmod"` // Дата останньої зміни
}

// ParseSitemap розбирає XML-дані sitemap
func ParseSitemap(data []byte) (interface{}, error) {
	// Спочатку пробуємо розпарсити як <urlset>
	var urlset URLSet
	if err := xml.Unmarshal(data, &urlset); err == nil && len(urlset.URLs) > 0 {
		return &urlset, nil
	}

	// Якщо не вийшло, пробуємо розпарсити як <sitemapindex>
	var sitemapIndex SitemapIndex
	if err := xml.Unmarshal(data, &sitemapIndex); err == nil && len(sitemapIndex.Sitemaps) > 0 {
		return &sitemapIndex, nil
	}

	return nil, fmt.Errorf("невідомий формат sitemap")
}
