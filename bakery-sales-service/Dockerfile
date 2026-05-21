FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY shared ./shared
COPY bakery-sales-service ./bakery-sales-service
WORKDIR /src/bakery-sales-service
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/bakery-sales-service ./cmd/server

FROM alpine:3.20
COPY --from=builder /bin/bakery-sales-service /bakery-sales-service
EXPOSE 50052
CMD ["/bakery-sales-service"]
