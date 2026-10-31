# Go QR Studio

A Go web application for scanning QR codes from images and generating QR codes from JSON or text.

## Features

- **Scan QR** — upload or drag and drop a PNG/JPEG image to read its QR code. Maximum file size: 10 MB.
- **Generate QR** — create a QR code from JSON or text, configure its size and error correction level, then download it as a PNG.
- English-language interface with Scan QR and Generate QR menus.
- The UI is embedded in the Go binary, so the app runs without separate frontend files.

## Run locally

Requirements: Go 1.21 or later.

```bash
go run .
```

Open [http://localhost:8080](http://localhost:8080).

Run the Go tests with:

```bash
go test ./...
```

## Run with Docker

Requirements: Docker with Docker Compose.

```bash
docker compose up --build
```

Open [http://localhost:8080](http://localhost:8080). Stop the app with:

```bash
docker compose down
```

The image uses a multi-stage build and a minimal runtime. The application listens on port `8080` inside the container.

## API endpoints

### Scan QR code

`POST /api/scan` accepts a multipart form with an `image` field. Supported image formats: PNG and JPEG.

Example:

```bash
curl -X POST http://localhost:8080/api/scan \
  -F "image=@qr.png"
```

Successful response:

```json
{"result":"QR code content"}
```

### Generate QR code

`POST /api/generate` accepts a JSON request. The `json` field contains the content to encode (as a string); `size` is optional and must be between 128 and 512 pixels; `level` is optional and can be `L`, `M`, `Q`, or `H`.

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"json":"{\"name\":\"Vostra\"}","size":256,"level":"M"}' \
  --output qr.png
```

On success, the endpoint returns a PNG image. If `size` or `level` is omitted, the app uses a size of 512 pixels and level `M`.

## Main project files

- `main.go` — HTTP server and API handlers.
- `go_qr_studio_ui.html` — UI page, embedded in the binary using `go:embed`.
- `Dockerfile` and `compose.yaml` — container configuration.
