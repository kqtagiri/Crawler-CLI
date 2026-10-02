# Crawler-CLI

CLI приложение для асинхронного обхода сайтов и построения дерева найденных страниц.

## Сборка

```bash
go build -o crawler-cli ./cmd/app
```

## Запуск

Через собранный бинарник:

```bash
./crawler-cli \
  --urls "https://go.dev,https://example.com" \
  --depth 2 \
  --timeout 2m \
  --request-timeout 10s \
  --output result.json \
  --log crawler.log
```

Без сборки:

```bash
go run ./cmd/app \
  --urls "https://go.dev,https://example.com" \
  --depth 2 \
  --timeout 2m \
  --request-timeout 10s \
  --output result.json \
  --log crawler.log
```

## Флаги

- `--urls` - Список стартовых URL через запятую
- `--depth` - Максимальная глубина рекурсивного обхода
- `--timeout` - Общий таймаут выполнения
- `--request-timeout` - Таймаут одного HTTP-запроса
- `--output` - Файл для JSON-результата
- `--log` - Файл для логов

## Тесты

Все тесты:

```bash
go test ./...
```

С проверкой гонок данных:

```bash
go test -race ./...
```

Повторные прогоны:

```bash
go test -race -count=20 ./...
```