package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	qrencode "github.com/skip2/go-qrcode"
)

const maxUploadSize = 10 << 20

//go:embed go_qr_studio_ui.html
var ui []byte

type apiResponse struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/api/scan", scanQR)
	mux.HandleFunc("/api/generate", generateQR)

	addr := ":8080"
	log.Printf("QR Scan & Generate berjalan di http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(ui); err != nil {
		log.Printf("gagal mengirim halaman UI: %v", err)
	}
}

func scanQR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Gunakan metode POST.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "Upload tidak valid atau ukuran file melebihi 10 MB.")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Pilih gambar QR terlebih dahulu.")
		return
	}
	defer file.Close()

	img, _, err := image.Decode(io.LimitReader(file, maxUploadSize))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Gambar tidak dapat dibaca. Gunakan format PNG atau JPEG.")
		return
	}

	bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Gambar tidak dapat diproses.")
		return
	}

	result, err := qrcode.NewQRCodeReader().Decode(bitmap, nil)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "QR code tidak ditemukan atau tidak dapat dibaca.")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Result: result.GetText()})
}

func generateQR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Gunakan metode POST.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var request struct {
		JSON  string `json:"json"`
		Size  int    `json:"size"`
		Level string `json:"level"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Isi dengan JSON yang valid.")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "Request harus berisi satu objek JSON.")
		return
	}
	if strings.TrimSpace(request.JSON) == "" {
		writeError(w, http.StatusBadRequest, "Konten QR tidak boleh kosong.")
		return
	}

	size := request.Size
	if size == 0 {
		size = 512
	}
	if size < 128 || size > 512 {
		writeError(w, http.StatusBadRequest, "Ukuran QR harus antara 128 dan 512 piksel.")
		return
	}

	level := qrencode.Medium
	switch strings.ToUpper(request.Level) {
	case "", "M":
	case "L":
		level = qrencode.Low
	case "Q":
		level = qrencode.High
	case "H":
		level = qrencode.Highest
	default:
		writeError(w, http.StatusBadRequest, "Level koreksi error harus L, M, Q, atau H.")
		return
	}

	png, err := qrencode.Encode(request.JSON, level, size)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "QR code gagal dibuat.")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `inline; filename="qr-code.png"`)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		log.Printf("gagal mengirim gambar QR: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiResponse{Error: strings.TrimSpace(message)})
}

func writeJSON(w http.ResponseWriter, status int, value apiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("gagal mengirim respons JSON: %v", fmt.Errorf("encode response: %w", err))
	}
}
