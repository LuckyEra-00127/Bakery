FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY shared ./shared
COPY bakery-management-service ./bakery-management-service
WORKDIR /src/bakery-management-service
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/bakery-management-service ./cmd/server

FROM alpine:3.20
COPY --from=builder /bin/bakery-management-service /bakery-management-service
EXPOSE 50053
CMD ["/bakery-management-service"]
