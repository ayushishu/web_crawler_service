FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY server/go.mod server/go.sum ./

# Download dependencies
RUN go mod tidy
COPY server/ .

RUN go build -o webcrawler .
FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /app/webcrawler .

EXPOSE 8080

# Command to run the application
CMD ["./webcrawler"]
