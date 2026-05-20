package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type gatewayCache struct {
	client *redis.Client
	ttl    time.Duration
}

func newGatewayCache(addr string) *gatewayCache {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		log.Println("redis cache disabled: REDIS_ADDR is empty")
		return nil
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("redis cache disabled: %v", err)
		return nil
	}
	log.Printf("redis cache connected: %s", addr)
	return &gatewayCache{client: client, ttl: 2 * time.Minute}
}

func (c *gatewayCache) Close() {
	if c != nil && c.client != nil {
		_ = c.client.Close()
	}
}

func (c *gatewayCache) Get(ctx context.Context, key string, dst any) (bool, error) {
	if c == nil || c.client == nil || key == "" {
		return false, nil
	}
	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(value), dst)
}

func (c *gatewayCache) Set(ctx context.Context, key string, value any) {
	if c == nil || c.client == nil || key == "" {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache marshal error: %v", err)
		return
	}
	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		log.Printf("cache set error: %v", err)
	}
}

func (c *gatewayCache) Delete(ctx context.Context, keys ...string) {
	if c == nil || c.client == nil || len(keys) == 0 {
		return
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache delete error: %v", err)
	}
}

func (c *gatewayCache) DeletePrefix(ctx context.Context, prefixes ...string) {
	if c == nil || c.client == nil || len(prefixes) == 0 {
		return
	}
	for _, prefix := range prefixes {
		iter := c.client.Scan(ctx, 0, prefix+"*", 100).Iterator()
		var keys []string
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			log.Printf("cache scan error: %v", err)
			continue
		}
		if len(keys) > 0 {
			c.Delete(ctx, keys...)
		}
	}
}

func (g *gateway) cachedInvokeHTTP(w http.ResponseWriter, r *http.Request, conn *grpc.ClientConn, method string, req any, res any, cacheKey string) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if ok, err := g.cache.Get(ctx, cacheKey, res); ok && err == nil {
		w.Header().Set("X-Cache", "HIT")
		respond(w, http.StatusOK, res)
		return
	} else if err != nil {
		log.Printf("cache get error for %s: %v", cacheKey, err)
	}

	if err := conn.Invoke(ctx, method, req, res); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": grpcErrorMessage(err)})
		return
	}

	g.cache.Set(ctx, cacheKey, res)
	w.Header().Set("X-Cache", "MISS")
	respond(w, http.StatusOK, res)
}

func (g *gateway) invalidateCatalogCache(ctx context.Context) {
	g.cache.DeletePrefix(ctx, "products:", "bake_plans:", "available_products:")
}

func (g *gateway) invalidateOrderCache(ctx context.Context) {
	g.cache.DeletePrefix(ctx, "orders:", "available_products:", "bake_plans:", "statistics:")
}
