# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/main.go
RUN go build -o encrypt cmd/secure/main.go
ENV DEFAULT_USER_PASS=yami

# Run stage
FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/encrypt .
COPY certs /app/certs
COPY frontend/static /app/frontend/static
COPY app.env .
COPY fapi_mock_response.json .
COPY start.sh .

RUN chmod +x start.sh

EXPOSE 8443

CMD [ "./main" ]

ENTRYPOINT [ "/app/start.sh" ]