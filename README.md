# ☀️ Solar Cast

**Solar Cast** estimates solar panel output for any location and panel type, using a Go backend, Vite + React frontend, and Redis caching.
Includes an scraper to collect and update panel data.

---

## 🚀 Features

Go API that combines scraped panel specifications with third-party irradiance/forecast APIs to calculate expected generation.

Location input: users enter an address/postcode; it’s geocoded to lat/long for accurate sun/irradiance calculations.

React frontend to query the API and visualise daily output.

Colly web scraper gathers panel specs from vendor pages and normalises them.

Redis caching for daily calculations to cut recomputation and external API calls.

---

## 🐳 Running with Docker

### 1. Build and start services

```bash
docker compose build
docker compose up -d
```

---

## 🧹 Optional Scraper

Run the scraper separately with:

```bash
docker compose --profile scraper up scraper
```

Adjust pages (example):

```bash
docker compose run scraper /usr/local/bin/solar-cast --scrape --pages 10
```

---

## 🔧 Development (without Docker)

### Backend

```bash
cd backend
go run main.go --server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

---
