package grpcdelivery

import (
	"context"

	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/domain"
	"github.com/bakeplan/bakeplan-go/bakery-management-service/internal/usecase"
	"github.com/bakeplan/bakeplan-go/shared/contracts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serviceName = "bakery.management.BakeryManagementService"

type Server struct {
	uc *usecase.ManagementUseCase
}

func NewServer(uc *usecase.ManagementUseCase) *Server {
	return &Server{uc: uc}
}

func Register(s *grpc.Server, server *Server) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*interface{})(nil),
		Methods: []grpc.MethodDesc{
			unary[contracts.CreateIngredientRequest]("CreateIngredient", (*Server).CreateIngredient),
			unary[contracts.UpdateIngredientRequest]("UpdateIngredient", unimplemented[contracts.UpdateIngredientRequest]),
			unary[contracts.DeleteIngredientRequest]("DeleteIngredient", unimplemented[contracts.DeleteIngredientRequest]),
			unary[contracts.ListIngredientsRequest]("ListIngredients", (*Server).ListIngredients),
			unary[contracts.CreateRecipeRequest]("CreateRecipe", unimplemented[contracts.CreateRecipeRequest]),
			unary[contracts.UpdateRecipeRequest]("UpdateRecipe", unimplemented[contracts.UpdateRecipeRequest]),
			unary[contracts.DeleteRecipeRequest]("DeleteRecipe", unimplemented[contracts.DeleteRecipeRequest]),
			unary[contracts.GetRecipeByProductIDRequest]("GetRecipeByProductId", unimplemented[contracts.GetRecipeByProductIDRequest]),
			unary[contracts.AddIngredientToRecipeRequest]("AddIngredientToRecipe", unimplemented[contracts.AddIngredientToRecipeRequest]),
			unary[contracts.RemoveIngredientFromRecipeRequest]("RemoveIngredientFromRecipe", unimplemented[contracts.RemoveIngredientFromRecipeRequest]),
			unary[contracts.CalculateRecipeRequest]("CalculateRecipe", unimplemented[contracts.CalculateRecipeRequest]),
			unary[contracts.CalculateRecipeCostRequest]("CalculateRecipeCost", unimplemented[contracts.CalculateRecipeCostRequest]),
			unary[contracts.CreateTaskRequest]("CreateTask", (*Server).CreateTask),
			unary[contracts.UpdateTaskRequest]("UpdateTask", (*Server).UpdateTask),
			unary[contracts.DeleteTaskRequest]("DeleteTask", (*Server).DeleteTask),
			unary[contracts.ListTasksRequest]("ListTasks", (*Server).ListTasks),
			unary[contracts.CompleteTaskRequest]("CompleteTask", (*Server).CompleteTask),
			unary[contracts.GetDailyStatisticsRequest]("GetDailyStatistics", (*Server).GetDailyStatistics),
			unary[contracts.GetWeeklyStatisticsRequest]("GetWeeklyStatistics", unimplemented[contracts.GetWeeklyStatisticsRequest]),
			unary[contracts.GetMonthlyStatisticsRequest]("GetMonthlyStatistics", unimplemented[contracts.GetMonthlyStatisticsRequest]),
			unary[contracts.SendEmailRequest]("SendEmail", (*Server).SendEmail),
			unary[contracts.ListEmailLogsRequest]("ListEmailLogs", (*Server).ListEmailLogs),
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

func (s *Server) CreateIngredient(ctx context.Context, req *contracts.CreateIngredientRequest) (any, error) {
	ingredient, err := s.uc.CreateIngredient(ctx, req.Name, req.Unit, req.CostPerUnit, toDomainSections(req.Sections))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.IngredientResponse{Ingredient: toIngredient(ingredient)}, nil
}

func (s *Server) ListIngredients(ctx context.Context, _ *contracts.ListIngredientsRequest) (any, error) {
	ingredients, err := s.uc.ListIngredients(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := make([]contracts.Ingredient, 0, len(ingredients))
	for _, i := range ingredients {
		response = append(response, toIngredient(i))
	}
	return &contracts.ListIngredientsResponse{Ingredients: response}, nil
}

func (s *Server) CreateTask(ctx context.Context, req *contracts.CreateTaskRequest) (any, error) {
	task, err := s.uc.CreateTask(ctx, req.Title, req.Status, req.DueDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.TaskResponse{Task: toTask(task)}, nil
}

func (s *Server) ListTasks(ctx context.Context, _ *contracts.ListTasksRequest) (any, error) {
	tasks, err := s.uc.ListTasks(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := make([]contracts.Task, 0, len(tasks))
	for _, t := range tasks {
		response = append(response, toTask(t))
	}
	return &contracts.ListTasksResponse{Tasks: response}, nil
}

func (s *Server) UpdateTask(ctx context.Context, req *contracts.UpdateTaskRequest) (any, error) {
	task, err := s.uc.UpdateTask(ctx, req.ID, req.Title, req.Status, req.DueDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.TaskResponse{Task: toTask(task)}, nil
}

func (s *Server) DeleteTask(ctx context.Context, req *contracts.DeleteTaskRequest) (any, error) {
	if err := s.uc.DeleteTask(ctx, req.ID); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &contracts.StatusResponse{Success: true, Message: "task deleted"}, nil
}

func (s *Server) CompleteTask(ctx context.Context, req *contracts.CompleteTaskRequest) (any, error) {
	task, err := s.uc.CompleteTask(ctx, req.ID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &contracts.TaskResponse{Task: toTask(task)}, nil
}

func (s *Server) GetDailyStatistics(ctx context.Context, req *contracts.GetDailyStatisticsRequest) (any, error) {
	stats, err := s.uc.ListDailyStatistics(ctx, req.Date)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := make([]contracts.DailyStatistic, 0, len(stats))
	for _, stat := range stats {
		response = append(response, contracts.DailyStatistic{Date: stat.Date, ProductID: stat.ProductID, Product: stat.Product, Baked: stat.Baked, Delivered: stat.Delivered, Sold: stat.Sold, Returned: stat.Returned, Left: stat.Left, Revenue: stat.Revenue})
	}
	return &contracts.DailyStatisticsResponse{Statistics: response}, nil
}

func (s *Server) SendEmail(ctx context.Context, req *contracts.SendEmailRequest) (any, error) {
	id, success, message, err := s.uc.SendEmail(ctx, req.To, req.Subject, req.Body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.SendEmailResponse{ID: id, Success: success, Message: message}, nil
}

func (s *Server) ListEmailLogs(ctx context.Context, _ *contracts.ListEmailLogsRequest) (any, error) {
	logs, err := s.uc.ListEmailLogs(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := make([]contracts.EmailLog, 0, len(logs))
	for _, logEntry := range logs {
		response = append(response, toEmailLog(logEntry))
	}
	return &contracts.ListEmailLogsResponse{EmailLogs: response}, nil
}

func toIngredient(ingredient domain.Ingredient) contracts.Ingredient {
	return contracts.Ingredient{ID: ingredient.ID, Name: ingredient.Name, Unit: ingredient.Unit, CostPerUnit: ingredient.CostPerUnit, Sections: toContractSections(ingredient.Sections)}
}

func toDomainSections(sections []contracts.IngredientSection) []domain.IngredientSection {
	out := make([]domain.IngredientSection, 0, len(sections))
	for _, section := range sections {
		out = append(out, domain.IngredientSection{Name: section.Name, Amount: section.Amount, Unit: section.Unit})
	}
	return out
}

func toContractSections(sections []domain.IngredientSection) []contracts.IngredientSection {
	out := make([]contracts.IngredientSection, 0, len(sections))
	for _, section := range sections {
		out = append(out, contracts.IngredientSection{Name: section.Name, Amount: section.Amount, Unit: section.Unit})
	}
	return out
}

func toTask(task domain.Task) contracts.Task {
	return contracts.Task{ID: task.ID, Title: task.Title, Status: task.Status, DueDate: task.DueDate, CreatedAt: task.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func toEmailLog(logEntry domain.EmailLog) contracts.EmailLog {
	return contracts.EmailLog{ID: logEntry.ID, Recipient: logEntry.Recipient, Subject: logEntry.Subject, Body: logEntry.Body, Status: logEntry.Status, CreatedAt: logEntry.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
}
