# Backend Build Stage
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o modern-fm-backend main.go

# Production Stage
FROM alpine:latest
RUN apk add --no-cache ffmpeg tzdata cifs-utils
WORKDIR /app
COPY --from=backend-builder /app/modern-fm-backend .
COPY index.html .

EXPOSE 38866
ENV GIN_MODE=release
CMD ["./modern-fm-backend"]
