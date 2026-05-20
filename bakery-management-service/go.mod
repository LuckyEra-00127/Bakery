module github.com/bakeplan/bakeplan-go/bakery-management-service

go 1.23

require (
    github.com/bakeplan/bakeplan-go/shared v0.0.0
    github.com/google/uuid v1.6.0
    github.com/lib/pq v1.10.9
    github.com/nats-io/nats.go v1.37.0
    google.golang.org/grpc v1.67.1
)

replace github.com/bakeplan/bakeplan-go/shared => ../shared
