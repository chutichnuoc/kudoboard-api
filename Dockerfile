FROM golang:1.24 AS builder
ARG VERSION=1.0.0
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X kudoboard-api/internal/api/handlers.Version=${VERSION}" -o application cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/application .
CMD ["./application"]
