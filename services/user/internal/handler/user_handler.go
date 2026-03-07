package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/ken/connect-microservice/gen/user/v1"
	"github.com/ken/connect-microservice/gen/user/v1/userv1connect"
	"github.com/ken/connect-microservice/services/user/internal/repository"
)

type UserHandler struct {
	repo *repository.UserRepository
}

var _ userv1connect.UserServiceHandler = (*UserHandler)(nil)

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *connect.Request[userv1.CreateUserRequest]) (*connect.Response[userv1.CreateUserResponse], error) {
	if req.Msg.Email == "" || req.Msg.Name == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("email, name and password are required"))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Msg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("hash password: %w", err))
	}

	role := "customer"
	if req.Msg.Role == userv1.Role_ROLE_ADMIN {
		role = "admin"
	}

	u, err := h.repo.Create(ctx, repository.User{
		Email:        req.Msg.Email,
		Name:         req.Msg.Name,
		Role:         role,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("create user: %w", err))
	}

	return connect.NewResponse(&userv1.CreateUserResponse{
		User: toProtoUser(u),
	}), nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *connect.Request[userv1.GetUserRequest]) (*connect.Response[userv1.GetUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	u, err := h.repo.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&userv1.GetUserResponse{
		User: toProtoUser(u),
	}), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *connect.Request[userv1.ListUsersRequest]) (*connect.Response[userv1.ListUsersResponse], error) {
	pageSize := int(req.Msg.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	page := int(req.Msg.Page)
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	users, total, err := h.repo.List(ctx, pageSize, offset)
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

	u, err := h.repo.Update(ctx, req.Msg.Id, req.Msg.Name, req.Msg.Email)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&userv1.UpdateUserResponse{
		User: toProtoUser(u),
	}), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *connect.Request[userv1.DeleteUserRequest]) (*connect.Response[userv1.DeleteUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	if err := h.repo.SoftDelete(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&userv1.DeleteUserResponse{}), nil
}

func toProtoUser(u repository.User) *userv1.User {
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
