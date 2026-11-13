# Нечёткие величины N(a, σ) — веб + Docker

Комплекс по практике «Неклассические логики»: библиотеки Go, REST API, минималистичный React UI.

## Запуск одной командой

```bash
cd fuzzy-normal
make
```

Собирается React, собирается Go-сервер, поднимается **http://localhost:8080** (UI + API).

Docker (альтернатива): `make up` · все команды: `make help`

Откройте в браузере: **http://localhost:8080**

Журнал операций сохраняется в volume `fuzzy-data` (`/app/data/fuzzy_journal.log` в контейнере).

## Локальная разработка

**Терминал 1:** `make dev-api`  
**Терминал 2:** `make dev-web`

UI: **http://localhost:5173**

## Структура

| Путь | Описание |
|------|----------|
| `pkg/fuzzy` | нечёткие величины, генерация, журнал |
| `pkg/cluster` | k-means |
| `cmd/server` | HTTP API + раздача React |
| `web` | React (Vite) |

Документация: [docs/README.md](docs/README.md).
