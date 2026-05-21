FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY shared ./shared
COPY user-service ./user-service
WORKDIR /src/user-service
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/user-service ./cmd/server

FROM alpine:3.20
COPY --from=builder /bin/user-service /user-service
EXPOSE 50051
CMD ["/user-service"]
