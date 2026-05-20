package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bakeplan/bakeplan-go/shared/contracts"
	"github.com/bakeplan/bakeplan-go/shared/grpcjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	userService       = "/bakery.user.UserService/"
	salesService      = "/bakery.sales.BakerySalesService/"
	managementService = "/bakery.management.BakeryManagementService/"
)

var (
	gatewayStartedAt       = time.Now()
	gatewayRequestsTotal   uint64
	gatewayErrorsTotal     uint64
	gatewayLatencyMicros   uint64
	gatewayHealthRequests  uint64
	gatewayMetricsRequests uint64
)

type gateway struct {
	userConn       *grpc.ClientConn
	salesConn      *grpc.ClientConn
	managementConn *grpc.ClientConn
	cache          *gatewayCache
}

func main() {
	grpcjson.Register()

	g := &gateway{
		userConn:       mustDial(env("USER_SERVICE_ADDR", "localhost:50051")),
		salesConn:      mustDial(env("SALES_SERVICE_ADDR", "localhost:50052")),
		managementConn: mustDial(env("MANAGEMENT_SERVICE_ADDR", "localhost:50053")),
		cache:          newGatewayCache(env("REDIS_ADDR", "localhost:6379")),
	}
	defer g.userConn.Close()
	defer g.salesConn.Close()
	defer g.managementConn.Close()
	defer g.cache.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&gatewayHealthRequests, 1)
		respond(w, http.StatusOK, map[string]any{"status": "ok", "service": "api-gateway"})
	})
	mux.HandleFunc("GET /metrics", metricsHandler)

	mux.HandleFunc("POST /auth/register", g.register)
	mux.HandleFunc("POST /auth/login", g.login)
	mux.HandleFunc("GET /auth/me", g.me)
	mux.HandleFunc("GET /admin/clients", g.listClients)

	mux.HandleFunc("POST /admin/products", g.createProduct)
	mux.HandleFunc("GET /products", g.listProducts)
	mux.HandleFunc("GET /admin/products", g.listProducts)

	mux.HandleFunc("POST /admin/bake-plans", g.createBakePlan)
	mux.HandleFunc("GET /bake-plans", g.listBakePlans)
	mux.HandleFunc("GET /available-products", g.availableProducts)

	mux.HandleFunc("POST /orders", g.createOrder)
	mux.HandleFunc("GET /orders/my", g.myOrders)
	mux.HandleFunc("GET /admin/orders", g.adminOrders)
	mux.HandleFunc("PATCH /admin/orders/status", g.updateOrderStatus)

	mux.HandleFunc("POST /admin/ingredients", g.createIngredient)
	mux.HandleFunc("GET /admin/ingredients", g.listIngredients)
	mux.HandleFunc("POST /admin/tasks", g.createTask)
	mux.HandleFunc("GET /admin/tasks", g.listTasks)
	mux.HandleFunc("PATCH /admin/tasks/done", g.completeTask)
	mux.HandleFunc("PATCH /admin/tasks/status", g.updateTaskStatus)
	mux.HandleFunc("DELETE /admin/tasks/", g.deleteTask)
	mux.HandleFunc("GET /admin/statistics/daily", g.dailyStatistics)
	mux.HandleFunc("POST /admin/email", g.sendEmail)
	mux.HandleFunc("GET /admin/email-logs", g.listEmailLogs)

	port := env("HTTP_PORT", "8080")
	log.Printf("api-gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, cors(metricsMiddleware(mux))); err != nil {
		log.Fatal(err)
	}
}

func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.DialContext(context.Background(), addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(grpcjson.Codec{})),
	)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func (g *gateway) register(w http.ResponseWriter, r *http.Request) {
	var req contracts.RegisterUserRequest
	if !decode(w, r, &req) {
		return
	}
	var res contracts.UserResponse
	g.invokeHTTP(w, r, g.userConn, userService+"RegisterUser", &req, &res)
}

func (g *gateway) login(w http.ResponseWriter, r *http.Request) {
	var req contracts.LoginUserRequest
	if !decode(w, r, &req) {
		return
	}
	var res contracts.LoginResponse
	g.invokeHTTP(w, r, g.userConn, userService+"LoginUser", &req, &res)
}

func (g *gateway) me(w http.ResponseWriter, r *http.Request) {
	user, ok := g.currentUser(w, r)
	if !ok {
		return
	}
	respond(w, http.StatusOK, map[string]any{"user": user})
}

func (g *gateway) listClients(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var res contracts.ListClientsResponse
	g.invokeHTTP(w, r, g.userConn, userService+"ListClients", &contracts.ListClientsRequest{}, &res)
}

func (g *gateway) createProduct(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.CreateProductRequest
	if !decode(w, r, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var res contracts.ProductResponse
	if err := g.salesConn.Invoke(ctx, salesService+"CreateProduct", &req, &res); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": grpcErrorMessage(err)})
		return
	}
	g.invalidateCatalogCache(ctx)
	respond(w, http.StatusOK, res)
}

func (g *gateway) listProducts(w http.ResponseWriter, r *http.Request) {
	var res contracts.ListProductsResponse
	g.cachedInvokeHTTP(w, r, g.salesConn, salesService+"ListProducts", &contracts.ListProductsRequest{}, &res, "products:list")
}

func (g *gateway) createBakePlan(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.CreateBakePlanRequest
	if !decode(w, r, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var res contracts.BakePlanResponse
	if err := g.salesConn.Invoke(ctx, salesService+"CreateBakePlan", &req, &res); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": grpcErrorMessage(err)})
		return
	}
	g.invalidateCatalogCache(ctx)
	respond(w, http.StatusOK, res)
}

func (g *gateway) listBakePlans(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	req := contracts.ListBakePlansByDateRequest{Date: date}
	var res contracts.ListBakePlansResponse
	g.cachedInvokeHTTP(w, r, g.salesConn, salesService+"ListBakePlansByDate", &req, &res, "bake_plans:"+date)
}

func (g *gateway) availableProducts(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	req := contracts.GetAvailableProductsRequest{Date: date}
	var res contracts.ListBakePlansResponse
	g.cachedInvokeHTTP(w, r, g.salesConn, salesService+"GetAvailableProducts", &req, &res, "available_products:"+date)
}

func (g *gateway) createOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := g.currentUser(w, r)
	if !ok {
		return
	}
	var req contracts.CreateOrderRequest
	if !decode(w, r, &req) {
		return
	}
	req.ClientID = user.ID
	if req.StoreName == "" {
		req.StoreName = user.FullName
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var res contracts.OrderResponse
	if err := g.salesConn.Invoke(ctx, salesService+"CreateOrder", &req, &res); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": grpcErrorMessage(err)})
		return
	}
	g.invalidateOrderCache(ctx)
	respond(w, http.StatusOK, res)
}

func (g *gateway) myOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := g.currentUser(w, r)
	if !ok {
		return
	}
	req := contracts.ListClientOrdersRequest{ClientID: user.ID}
	var res contracts.ListOrdersResponse
	g.invokeHTTP(w, r, g.salesConn, salesService+"ListClientOrders", &req, &res)
}

func (g *gateway) adminOrders(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var res contracts.ListOrdersResponse
	g.invokeHTTP(w, r, g.salesConn, salesService+"ListAdminOrders", &contracts.ListAdminOrdersRequest{}, &res)
}

func (g *gateway) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.UpdateOrderStatusRequest
	if !decode(w, r, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var res contracts.OrderResponse
	if err := g.salesConn.Invoke(ctx, salesService+"UpdateOrderStatus", &req, &res); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": grpcErrorMessage(err)})
		return
	}
	g.invalidateOrderCache(ctx)
	respond(w, http.StatusOK, res)
}

func (g *gateway) ensureProductFromIngredient(ctx context.Context, ingredient contracts.Ingredient) error {
	name := strings.TrimSpace(ingredient.Name)
	if name == "" {
		return nil
	}

	var products contracts.ListProductsResponse
	if err := g.salesConn.Invoke(ctx, salesService+"ListProducts", &contracts.ListProductsRequest{}, &products); err != nil {
		return err
	}
	for _, product := range products.Products {
		if strings.EqualFold(strings.TrimSpace(product.Name), name) {
			return nil
		}
	}

	var productRes contracts.ProductResponse
	return g.salesConn.Invoke(ctx, salesService+"CreateProduct", &contracts.CreateProductRequest{
		Name:  name,
		Price: ingredient.CostPerUnit,
	}, &productRes)
}

func (g *gateway) createIngredient(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.CreateIngredientRequest
	if !decode(w, r, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var res contracts.IngredientResponse
	if err := g.managementConn.Invoke(ctx, managementService+"CreateIngredient", &req, &res); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": grpcErrorMessage(err)})
		return
	}
	if err := g.ensureProductFromIngredient(ctx, res.Ingredient); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": "ingredient was saved, but product sync failed: " + grpcErrorMessage(err)})
		return
	}
	g.invalidateCatalogCache(ctx)
	respond(w, http.StatusOK, res)
}

func (g *gateway) listIngredients(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var res contracts.ListIngredientsResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"ListIngredients", &contracts.ListIngredientsRequest{}, &res)
}

func (g *gateway) createTask(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.CreateTaskRequest
	if !decode(w, r, &req) {
		return
	}
	var res contracts.TaskResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"CreateTask", &req, &res)
}

func (g *gateway) listTasks(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var res contracts.ListTasksResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"ListTasks", &contracts.ListTasksRequest{}, &res)
}

func (g *gateway) completeTask(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.CompleteTaskRequest
	if !decode(w, r, &req) {
		return
	}
	var res contracts.TaskResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"CompleteTask", &req, &res)
}

func (g *gateway) dailyStatistics(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	date := r.URL.Query().Get("date")
	req := contracts.GetDailyStatisticsRequest{Date: date}
	var res contracts.DailyStatisticsResponse
	g.cachedInvokeHTTP(w, r, g.managementConn, managementService+"GetDailyStatistics", &req, &res, "statistics:daily:"+date)
}

func (g *gateway) sendEmail(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.SendEmailRequest
	if !decode(w, r, &req) {
		return
	}
	var res contracts.SendEmailResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"SendEmail", &req, &res)
}

func (g *gateway) listEmailLogs(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var res contracts.ListEmailLogsResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"ListEmailLogs", &contracts.ListEmailLogsRequest{}, &res)
}

func (g *gateway) invokeHTTP(w http.ResponseWriter, r *http.Request, conn *grpc.ClientConn, method string, req any, res any) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if err := conn.Invoke(ctx, method, req, res); err != nil {
		respond(w, grpcHTTPStatus(err), map[string]any{"error": grpcErrorMessage(err)})
		return
	}
	respond(w, http.StatusOK, res)
}

func (g *gateway) updateTaskStatus(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	var req contracts.UpdateTaskRequest
	if !decode(w, r, &req) {
		return
	}
	var res contracts.TaskResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"UpdateTask", &req, &res)
}

func (g *gateway) deleteTask(w http.ResponseWriter, r *http.Request) {
	if !g.requireAdmin(w, r) {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/tasks/")
	if id == "" || id == r.URL.Path {
		respond(w, http.StatusBadRequest, map[string]any{"error": "task id is required"})
		return
	}
	var res contracts.StatusResponse
	g.invokeHTTP(w, r, g.managementConn, managementService+"DeleteTask", &contracts.DeleteTaskRequest{ID: id}, &res)
}

func (g *gateway) currentUser(w http.ResponseWriter, r *http.Request) (contracts.User, bool) {
	user, ok := g.currentUserSilent(r)
	if !ok {
		respond(w, http.StatusUnauthorized, map[string]any{"error": "missing or invalid Authorization header"})
		return contracts.User{}, false
	}
	return user, true
}

func (g *gateway) currentUserSilent(r *http.Request) (contracts.User, bool) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return contracts.User{}, false
	}
	var res contracts.ValidateTokenResponse
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err := g.userConn.Invoke(ctx, userService+"ValidateToken", &contracts.ValidateTokenRequest{Token: token}, &res)
	if err != nil || !res.Valid {
		return contracts.User{}, false
	}
	return res.User, true
}

func (g *gateway) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	user, ok := g.currentUser(w, r)
	if !ok {
		return false
	}
	if user.Role != "ADMIN" {
		respond(w, http.StatusForbidden, map[string]any{"error": "admin access is required"})
		return false
	}
	return true
}

func grpcHTTPStatus(err error) int {
	s, ok := status.FromError(err)
	if !ok {
		return http.StatusBadGateway
	}
	switch s.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.FailedPrecondition, codes.ResourceExhausted:
		return http.StatusConflict
	case codes.Unavailable, codes.DeadlineExceeded:
		return http.StatusServiceUnavailable
	default:
		return http.StatusBadGateway
	}
}

func grpcErrorMessage(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	return err.Error()
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil {
		respond(w, http.StatusBadRequest, map[string]any{"error": "request body is required"})
		return false
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respond(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return false
	}
	return true
}

func respond(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		atomic.AddUint64(&gatewayRequestsTotal, 1)
		atomic.AddUint64(&gatewayLatencyMicros, uint64(time.Since(start).Microseconds()))
		if recorder.status >= http.StatusBadRequest {
			atomic.AddUint64(&gatewayErrorsTotal, 1)
		}
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&gatewayMetricsRequests, 1)
	requests := atomic.LoadUint64(&gatewayRequestsTotal)
	errors := atomic.LoadUint64(&gatewayErrorsTotal)
	latencyMicros := atomic.LoadUint64(&gatewayLatencyMicros)
	uptime := time.Since(gatewayStartedAt).Seconds()
	avgLatency := 0.0
	if requests > 0 {
		avgLatency = float64(latencyMicros) / float64(requests) / 1_000_000
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = fmt.Fprintf(w, "# HELP bakeplan_gateway_requests_total Total HTTP requests handled by API Gateway\n")
	_, _ = fmt.Fprintf(w, "# TYPE bakeplan_gateway_requests_total counter\n")
	_, _ = fmt.Fprintf(w, "bakeplan_gateway_requests_total %d\n", requests)
	_, _ = fmt.Fprintf(w, "# HELP bakeplan_gateway_errors_total Total HTTP requests with status code 400 or greater\n")
	_, _ = fmt.Fprintf(w, "# TYPE bakeplan_gateway_errors_total counter\n")
	_, _ = fmt.Fprintf(w, "bakeplan_gateway_errors_total %d\n", errors)
	_, _ = fmt.Fprintf(w, "# HELP bakeplan_gateway_latency_seconds_sum Total gateway request latency in seconds\n")
	_, _ = fmt.Fprintf(w, "# TYPE bakeplan_gateway_latency_seconds_sum counter\n")
	_, _ = fmt.Fprintf(w, "bakeplan_gateway_latency_seconds_sum %.6f\n", float64(latencyMicros)/1_000_000)
	_, _ = fmt.Fprintf(w, "# HELP bakeplan_gateway_latency_seconds_avg Average gateway request latency in seconds\n")
	_, _ = fmt.Fprintf(w, "# TYPE bakeplan_gateway_latency_seconds_avg gauge\n")
	_, _ = fmt.Fprintf(w, "bakeplan_gateway_latency_seconds_avg %.6f\n", avgLatency)
	_, _ = fmt.Fprintf(w, "# HELP bakeplan_gateway_uptime_seconds Gateway process uptime in seconds\n")
	_, _ = fmt.Fprintf(w, "# TYPE bakeplan_gateway_uptime_seconds gauge\n")
	_, _ = fmt.Fprintf(w, "bakeplan_gateway_uptime_seconds %.0f\n", uptime)
	_, _ = fmt.Fprintf(w, "# HELP bakeplan_gateway_health_requests_total Health endpoint requests\n")
	_, _ = fmt.Fprintf(w, "# TYPE bakeplan_gateway_health_requests_total counter\n")
	_, _ = fmt.Fprintf(w, "bakeplan_gateway_health_requests_total %d\n", atomic.LoadUint64(&gatewayHealthRequests))
	_, _ = fmt.Fprintf(w, "# HELP bakeplan_gateway_metrics_requests_total Metrics endpoint requests\n")
	_, _ = fmt.Fprintf(w, "# TYPE bakeplan_gateway_metrics_requests_total counter\n")
	_, _ = fmt.Fprintf(w, "bakeplan_gateway_metrics_requests_total %d\n", atomic.LoadUint64(&gatewayMetricsRequests))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if strings.EqualFold(r.Method, http.MethodOptions) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
