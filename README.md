# Scraper

This is a go based web scraper service, which ouputs clean markdown content of the webpage and screenshots, through an easy to use REST API.

The project now also includes an MCP Server which launches a HTTP MCP server

Scraper can automatically handle cookie consent banners (cookie consent rules from [here](https://github.com/duckduckgo/autoconsent))

## Prerequisites

- Go 1.21 or higher
- Chrome/Chromium browser installed on your system
- Git (for cloning the repository)

## Building

### Clone the Repository
```bash
git clone https://github.com/SubhanAfz/crawler.git
cd crawler
```

### Download Dependencies
```bash
# Download and install all Go dependencies
go mod download
go mod tidy
```

### Build the REST API Binary
```bash
# Build the binary (outputs to bin/)
make build
```

### Build the MCP Server Binary
```bash
make build-mcp
```
###
To run the binaries, you must include the rules.json, in the same folder as the executable.
The built binary will be located at `bin/` and can be executed directly.

## REST API Documentation

The scraper service runs on `http://localhost:8080` by default and provides two main endpoints:

### 1. Get Page Content

**Endpoint:** `GET /get_page`

**Description:** Scrapes a webpage, automatically handles cookie consent banners, and returns the page content with optional format conversion.

**Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `url` | string | Yes | - | The URL of the webpage to scrape |
| `wait_time` | integer | No | 1000 | Wait time in milliseconds after page load |
| `format` | string | No | - | Output format conversion (`markdown`) |


**Response:**
```json
{
  "title": "Page Title",
  "content": "Page content...",
  "url": "https://example.com"
}
```

### 2. Take Screenshot

**Endpoint:** `GET /screenshot`

**Description:** Takes a full-page screenshot of a webpage after automatically handling cookie consent banners.

**Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `url` | string | Yes | - | The URL of the webpage to capture |
| `wait_time` | integer | No | 1000 | Wait time in milliseconds after page load |

**Response:**
```json
{
  "image": "base64-encoded-image-data..."
}
```

### Error Responses

All endpoints return standardized error responses:

**Format:**
```json
{
  "error": "Error description"
}
```

**HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (missing/invalid parameters)
- `500` - Internal Server Error (scraping/processing failed)