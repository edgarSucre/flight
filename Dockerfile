# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/main.go

# Run stage
FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/main .
COPY certs /app/certs
COPY frontend/static /app/frontend/static
COPY app.env .
COPY fapi_mock_response.json .

EXPOSE 8443

ENTRYPOINT [ "./main" ]