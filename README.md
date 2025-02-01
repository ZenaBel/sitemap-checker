# Sitemap Checker

Sitemap Checker — це інструмент для перевірки сторінок вебсайту на основі `sitemap.xml`. Він допомагає виявляти помилки, такі як неправильні статус-коди, редіректи, відсутність канонічних посилань, дублі контенту та інші SEO-проблеми. Результати перевірки зберігаються у JSON-файлі для подальшого аналізу.

## Run Locally

Щоб запустити проєкт локально, виконайте наступні кроки:

1. Клонуйте репозиторій:
   ```bash
   git clone https://github.com/ZenaBel/sitemap-checker.git
   cd sitemap-checker
   ```

2. Встановіть залежності:
   ```bash
   go mod download
   ```

3. Запустіть додаток:
   ```bash
   make run
   ```

Або використовуйте Docker:

1. Зберіть Docker-образ:
   ```bash
   make docker-build
   ```

2. Запустіть контейнери:
   ```bash
   make docker-up
   ```

3. Перегляньте логи контейнера:
   ```bash
   make logs
   ```

4. Зупиніть контейнери:
   ```bash
   make docker-down
   ```

5. Очистіть проєкт (видаліть бінарники, логи та контейнери):
   ```bash
   make clean
   ```

Для отримання додаткової інформації використовуйте:
```bash
make help
```

## Usage/Examples

Після запуску додаток проаналізує `sitemap.xml` і перевірить кожну сторінку. Результати будуть збережені у файлі `results.json`. Приклад виводу:

```json
[
  {
    "url": "https://example.com/page1",
    "status_code": 200,
    "redirects": [],
    "canonical_url": "https://example.com/canonical-page1",
    "meta_tags": {
      "title": "Example Page 1",
      "description": "This is an example page."
    },
    "load_time": "1.23s",
    "is_blocked_by_robots_txt": false,
    "content_hash": "a1b2c3d4e5f6..."
  }
]
```

## Environment Variables

Для налаштування додатка використовуйте змінні середовища. Створіть файл `.env` у корені проєкту з наступним вмістом:

```env
# Налаштування додатка
SITEMAP_URL=https://example.com/sitemap.xml
TIMEOUT=30s
MAX_GOROUTINES=10
MAX_DEPTH=10
MAX_REDIRECTS=5

# Налаштування Redis
REDIS_URL=redis:6379
```

## License

Цей проєкт ліцензовано за умовами [MIT License](https://choosealicense.com/licenses/mit/).