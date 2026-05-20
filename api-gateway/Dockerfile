FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY shared ./shared
COPY api-gateway ./api-gateway
WORKDIR /src/api-gateway
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api-gateway ./cmd

FROM alpine:3.20
COPY --from=builder /bin/api-gateway /api-gateway
EXPOSE 8080
CMD ["/api-gateway"]
