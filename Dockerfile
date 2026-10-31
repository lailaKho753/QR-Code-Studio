FROM golang:1.21-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go go_qr_studio_ui.html ./
RUN go test ./... && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/qr-scan-generate .

FROM scratch

COPY --from=builder /out/qr-scan-generate /qr-scan-generate

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/qr-scan-generate"]
