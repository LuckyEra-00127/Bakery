package grpcdelivery

import (
	"context"
	"strings"

	"github.com/bakeplan/bakeplan-go/shared/contracts"
	"github.com/bakeplan/bakeplan-go/user-service/internal/domain"
	"github.com/bakeplan/bakeplan-go/user-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serviceName = "bakery.user.UserService"

type Server struct {
	uc *usecase.UserUseCase
}

func NewServer(uc *usecase.UserUseCase) *Server {
	return &Server{uc: uc}
}

func Register(s *grpc.Server, server *Server) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*interface{})(nil),
		Methods: []grpc.MethodDesc{
			unary[contracts.RegisterUserRequest]("RegisterUser", (*Server).RegisterUser),
			unary[contracts.LoginUserRequest]("LoginUser", (*Server).LoginUser),
			unary[contracts.ValidateTokenRequest]("ValidateToken", (*Server).ValidateToken),
			unary[contracts.RefreshTokenRequest]("RefreshToken", unimplemented[contracts.RefreshTokenRequest]),
			unary[contracts.GetUserByIDRequest]("GetUserById", (*Server).GetUserByID),
			unary[contracts.GetUserByEmailRequest]("GetUserByEmail", (*Server).GetUserByEmail),
			unary[contracts.UpdateUserProfileRequest]("UpdateUserProfile", unimplemented[contracts.UpdateUserProfileRequest]),
			unary[contracts.ChangePasswordRequest]("ChangePassword", unimplemented[contracts.ChangePasswordRequest]),
			unary[contracts.CreateStoreProfileRequest]("CreateStoreProfile", unimplemented[contracts.CreateStoreProfileRequest]),
			unary[contracts.UpdateStoreProfileRequest]("UpdateStoreProfile", unimplemented[contracts.UpdateStoreProfileRequest]),
			unary[contracts.ListClientsRequest]("ListClients", (*Server).ListClients),
			unary[contracts.AssignRoleRequest]("AssignRole", unimplemented[contracts.AssignRoleRequest]),
			unary[contracts.DeleteUserRequest]("DeleteUser", (*Server).DeleteUser),
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

func (s *Server) RegisterUser(ctx context.Context, req *contracts.RegisterUserRequest) (any, error) {
	user, token, err := s.uc.Register(ctx, req.Email, req.Password, req.FullName, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &contracts.UserResponse{User: toContract(user), Token: token, Message: "registered"}, nil
}

func (s *Server) LoginUser(ctx context.Context, req *contracts.LoginUserRequest) (any, error) {
	user, token, err := s.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &contracts.LoginResponse{User: toContract(user), AccessToken: token}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *contracts.ValidateTokenRequest) (any, error) {
	token := strings.TrimPrefix(req.Token, "Bearer ")
	user, err := s.uc.ValidateToken(ctx, token)
	if err != nil {
		return &contracts.ValidateTokenResponse{Valid: false, Error: err.Error()}, nil
	}
	return &contracts.ValidateTokenResponse{Valid: true, User: toContract(user)}, nil
}

func (s *Server) GetUserByID(ctx context.Context, req *contracts.GetUserByIDRequest) (any, error) {
	user, err := s.uc.GetByID(ctx, req.ID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &contracts.UserResponse{User: toContract(user)}, nil
}

func (s *Server) GetUserByEmail(ctx context.Context, req *contracts.GetUserByEmailRequest) (any, error) {
	user, err := s.uc.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &contracts.UserResponse{User: toContract(user)}, nil
}

func (s *Server) ListClients(ctx context.Context, _ *contracts.ListClientsRequest) (any, error) {
	users, err := s.uc.ListClients(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	clients := make([]contracts.User, 0, len(users))
	for _, user := range users {
		clients = append(clients, toContract(user))
	}
	return &contracts.ListClientsResponse{Clients: clients}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *contracts.DeleteUserRequest) (any, error) {
	if err := s.uc.Delete(ctx, req.ID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &contracts.StatusResponse{Success: true, Message: "user deleted"}, nil
}

func toContract(user domain.User) contracts.User {
	return contracts.User{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
