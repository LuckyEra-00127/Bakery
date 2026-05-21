package grpcdelivery

import (
	"context"
	"strings"

	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/domain"
	"github.com/bakeplan/bakeplan-go/bakery-sales-service/internal/usecase"
	"github.com/bakeplan/bakeplan-go/shared/contracts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serviceName = "bakery.sales.BakerySalesService"

type Server struct {
	uc *usecase.SalesUseCase
}

func NewServer(uc *usecase.SalesUseCase) *Server {
	return &Server{uc: uc}
}

func Register(s *grpc.Server, server *Server) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*interface{})(nil),
		Methods: []grpc.MethodDesc{
			unary[contracts.CreateProductRequest]("CreateProduct", (*Server).CreateProduct),
			unary[contracts.UpdateProductRequest]("UpdateProduct", unimplemented[contracts.UpdateProductRequest]),
			unary[contracts.DeleteProductRequest]("DeleteProduct", unimplemented[contracts.DeleteProductRequest]),
			unary[contracts.GetProductByIDRequest]("GetProductById", unimplemented[contracts.GetProductByIDRequest]),
			unary[contracts.ListProductsRequest]("ListProducts", (*Server).ListProducts),
			unary[contracts.CreateBakePlanRequest]("CreateBakePlan", (*Server).CreateBakePlan),
			unary[contracts.UpdateBakePlanRequest]("UpdateBakePlan", unimplemented[contracts.UpdateBakePlanRequest]),
			unary[contracts.DeleteBakePlanRequest]("DeleteBakePlan", unimplemented[contracts.DeleteBakePlanRequest]),
			unary[contracts.ListBakePlansByDateRequest]("ListBakePlansByDate", (*Server).ListBakePlansByDate),
			unary[contracts.GetAvailableProductsRequest]("GetAvailableProducts", (*Server).GetAvailableProducts),
			unary[contracts.CreateOrderRequest]("CreateOrder", (*Server).CreateOrder),
			unary[contracts.CancelOrderRequest]("CancelOrder", unimplemented[contracts.CancelOrderRequest]),
			unary[contracts.GetOrderByIDRequest]("GetOrderById", unimplemented[contracts.GetOrderByIDRequest]),
			unary[contracts.UpdateOrderStatusRequest]("UpdateOrderStatus", (*Server).UpdateOrderStatus),
			unary[contracts.ListClientOrdersRequest]("ListClientOrders", (*Server).ListClientOrders),
			unary[contracts.ListAdminOrdersRequest]("ListAdminOrders", (*Server).ListAdminOrders),
			unary[contracts.CreateDeliveryRequest]("CreateDelivery", unimplemented[contracts.CreateDeliveryRequest]),
			unary[contracts.GetDeliveryByIDRequest]("GetDeliveryById", unimplemented[contracts.GetDeliveryByIDRequest]),
			unary[contracts.ListDeliveriesRequest]("ListDeliveries", unimplemented[contracts.ListDeliveriesRequest]),
			unary[contracts.CreateReturnReportRequest]("CreateReturnReport", unimplemented[contracts.CreateReturnReportRequest]),
			unary[contracts.UpdateReturnReportRequest]("UpdateReturnReport", unimplemented[contracts.UpdateReturnReportRequest]),
			unary[contracts.ListReturnReportsRequest]("ListReturnReports", unimplemented[contracts.ListReturnReportsRequest]),
			unary[contracts.CalculateRevenueRequest]("CalculateRevenue", (*Server).CalculateRevenue),
		},
	}, server)
}

func unary[Req any](method string, fn func(*Server, context.Context, *Req) (any, error)) grpc.MethodDesc {
	return grpc.MethodDesc{
		MethodName: method,
		Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
			in := new(Req)
			if err := dec(in); err != nil {
				return nil, err
			}
			server := srv.(*Server)
			if interceptor == nil {
				return fn(server, ctx, in)
			}
			info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/" + method}
			handler := func(ctx context.Context, req any) (any, error) {
				return fn(server, ctx, req.(*Req))
			}
			return interceptor(ctx, in, info, handler)
		},
	}
}

func unimplemented[Req any](_ *Server, _ context.Context, _ *Req) (any, error) {
	return nil, status.Error(codes.Unimplemented, "method is declared but not implemented yet")
}

func (s *Server) CreateProduct(ctx context.Context, req *contracts.CreateProductRequest) (any, error) {
	product, err := s.uc.CreateProduct(ctx, req.Name, req.Price)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.ProductResponse{Product: toProduct(product)}, nil
}

func (s *Server) ListProducts(ctx context.Context, _ *contracts.ListProductsRequest) (any, error) {
	products, err := s.uc.ListProducts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := make([]contracts.Product, 0, len(products))
	for _, p := range products {
		response = append(response, toProduct(p))
	}
	return &contracts.ListProductsResponse{Products: response}, nil
}

func (s *Server) CreateBakePlan(ctx context.Context, req *contracts.CreateBakePlanRequest) (any, error) {
	plan, err := s.uc.CreateBakePlan(ctx, req.ProductID, req.PlanDate, req.PlannedQuantity)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.BakePlanResponse{BakePlan: toBakePlan(plan)}, nil
}

func (s *Server) ListBakePlansByDate(ctx context.Context, req *contracts.ListBakePlansByDateRequest) (any, error) {
	plans, err := s.uc.ListBakePlansByDate(ctx, req.Date)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &contracts.ListBakePlansResponse{BakePlans: toBakePlans(plans)}, nil
}

func (s *Server) GetAvailableProducts(ctx context.Context, req *contracts.GetAvailableProductsRequest) (any, error) {
	plans, err := s.uc.ListBakePlansByDate(ctx, req.Date)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	available := make([]domain.BakePlan, 0, len(plans))
	for _, p := range plans {
		if p.AvailableQuantity > 0 {
			available = append(available, p)
		}
	}
	return &contracts.ListBakePlansResponse{BakePlans: toBakePlans(available)}, nil
}

func (s *Server) CreateOrder(ctx context.Context, req *contracts.CreateOrderRequest) (any, error) {
	order, err := s.uc.CreateOrder(ctx, req.ClientID, req.StoreName, req.BakePlanID, req.Quantity)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not enough") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.OrderResponse{Order: toOrder(order)}, nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context, req *contracts.UpdateOrderStatusRequest) (any, error) {
	order, err := s.uc.UpdateOrderStatus(ctx, req.ID, req.Status)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "not found") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.OrderResponse{Order: toOrder(order)}, nil
}

func (s *Server) ListClientOrders(ctx context.Context, req *contracts.ListClientOrdersRequest) (any, error) {
	orders, err := s.uc.ListClientOrders(ctx, req.ClientID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &contracts.ListOrdersResponse{Orders: toOrders(orders)}, nil
}

func (s *Server) ListAdminOrders(ctx context.Context, _ *contracts.ListAdminOrdersRequest) (any, error) {
	orders, err := s.uc.ListClientOrders(ctx, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &contracts.ListOrdersResponse{Orders: toOrders(orders)}, nil
}

func (s *Server) CalculateRevenue(_ context.Context, req *contracts.CalculateRevenueRequest) (any, error) {
	return &contracts.CalculateRevenueResponse{Revenue: s.uc.CalculateRevenue(req.SoldQuantity, req.ProductPrice)}, nil
}

func toProduct(product domain.Product) contracts.Product {
	return contracts.Product{ID: product.ID, Name: product.Name, Price: product.Price, CreatedAt: product.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func toBakePlan(plan domain.BakePlan) contracts.BakePlan {
	return contracts.BakePlan{ID: plan.ID, ProductID: plan.ProductID, ProductName: plan.ProductName, PlanDate: plan.PlanDate, PlannedQuantity: plan.PlannedQuantity, AvailableQuantity: plan.AvailableQuantity}
}

func toBakePlans(plans []domain.BakePlan) []contracts.BakePlan {
	response := make([]contracts.BakePlan, 0, len(plans))
	for _, plan := range plans {
		response = append(response, toBakePlan(plan))
	}
	return response
}

func toOrder(order domain.Order) contracts.Order {
	items := make([]contracts.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, contracts.OrderItem{ID: item.ID, OrderID: item.OrderID, BakePlanID: item.BakePlanID, ProductID: item.ProductID, Quantity: item.Quantity, UnitPrice: item.UnitPrice})
	}
	return contracts.Order{ID: order.ID, ClientID: order.ClientID, StoreName: order.StoreName, Status: order.Status, Items: items, CreatedAt: order.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func toOrders(orders []domain.Order) []contracts.Order {
	response := make([]contracts.Order, 0, len(orders))
	for _, order := range orders {
		response = append(response, toOrder(order))
	}
	return response
}
