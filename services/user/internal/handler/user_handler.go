package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/ken/connect-microservice/gen/user/v1"
	"github.com/ken/connect-microservice/gen/user/v1/userv1connect"
	"github.com/ken/connect-microservice/internal/auth"
	"github.com/ken/connect-microservice/services/user/internal/domain"
	"github.com/ken/connect-microservice/services/user/internal/usecase"
)

type UserHandler struct {
	uc       *usecase.UserUsecase
	tokenGen *auth.TokenGenerator
}

var _ userv1connect.UserServiceHandler = (*UserHandler)(nil)

func NewUserHandler(uc *usecase.UserUsecase, tokenGen *auth.TokenGenerator) *UserHandler {
	return &UserHandler{uc: uc, tokenGen: tokenGen}
}

func (h *UserHandler) Login(ctx context.Context, req *connect.Request[userv1.LoginRequest]) (*connect.Response[userv1.LoginResponse], error) {
	if req.Msg.Email == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("email and password are required"))
	}

	u, err := h.uc.Authenticate(ctx, req.Msg.Email, req.Msg.Password)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	token, expiresAt, err := h.tokenGen.GenerateToken(u.ID, u.Role)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("generate token: %w", err))
	}

	return connect.NewResponse(&userv1.LoginResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt.Unix(),
	}), nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *connect.Request[userv1.CreateUserRequest]) (*connect.Response[userv1.CreateUserResponse], error) {
	if req.Msg.Email == "" || req.Msg.Name == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("email, name and password are required"))
	}

	role := "customer"
	if req.Msg.Role == userv1.Role_ROLE_ADMIN {
		role = "admin"
	}

	u, err := h.uc.CreateUser(ctx, req.Msg.Email, req.Msg.Name, req.Msg.Password, role)
	if err != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("create user: %w", err))
	}
	return connect.NewResponse(&userv1.CreateUserResponse{User: toProtoUser(u)}), nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *connect.Request[userv1.GetUserRequest]) (*connect.Response[userv1.GetUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	u, err := h.uc.GetUser(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&userv1.GetUserResponse{User: toProtoUser(u)}), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *connect.Request[userv1.ListUsersRequest]) (*connect.Response[userv1.ListUsersResponse], error) {
	users, total, err := h.uc.ListUsers(ctx, int(req.Msg.PageSize), int(req.Msg.Page))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoUsers := make([]*userv1.User, len(users))
	for i, u := range users {
		protoUsers[i] = toProtoUser(u)
	}
	return connect.NewResponse(&userv1.ListUsersResponse{
		Users:      protoUsers,
		TotalCount: int32(total),
	}), nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *connect.Request[userv1.UpdateUserRequest]) (*connect.Response[userv1.UpdateUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	u, err := h.uc.UpdateUser(ctx, req.Msg.Id, req.Msg.Name, req.Msg.Email)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&userv1.UpdateUserResponse{User: toProtoUser(u)}), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *connect.Request[userv1.DeleteUserRequest]) (*connect.Response[userv1.DeleteUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	if err := h.uc.DeleteUser(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&userv1.DeleteUserResponse{}), nil
}

func toProtoUser(u domain.User) *userv1.User {
	role := userv1.Role_ROLE_CUSTOMER
	if u.Role == "admin" {
		role = userv1.Role_ROLE_ADMIN
	}
	return &userv1.User{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      role,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
