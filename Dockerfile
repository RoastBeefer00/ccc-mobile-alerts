FROM golang:latest AS builder
RUN mkdir /app
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
RUN chmod +x /server
EXPOSE 3000
CMD ["/server"]
