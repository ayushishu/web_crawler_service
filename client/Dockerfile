FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY client/ .
RUN go mod tidy
RUN go build -o client-ui .

#Create a minimal image for running the Go application
FROM alpine:latest
# Set the working directory
WORKDIR /root/
COPY --from=builder /app/client-ui .

EXPOSE 3000
CMD ["./client-ui"]
