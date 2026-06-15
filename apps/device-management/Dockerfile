FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o device-mgmt ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/device-mgmt .

EXPOSE 8080

CMD ["./device-mgmt"]
