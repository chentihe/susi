package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"strconv"

	"github.com/tihe/susi-auth-service/internal/model"
	"github.com/tihe/susi-auth-service/internal/service"
	"github.com/tihe/susi-proto/auth"
	"github.com/tihe/susi-proto/common"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *AuthHandler) createErrorResponse(message string, errorCode common.ErrorCode) *common.ResponseStatus {
	return &common.ResponseStatus{
		Success:   false,
		Message:   message,
		ErrorCode: errorCode.String(),
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	log.Printf("Register called with: name=%s, email=%s", req.Name, req.Email)
	role := h.protoRoleToModel(req.Role)

	secret, err := service.GenerateTOTPSecret(req.Name)
	if err != nil {
		log.Printf("Error generating TOTP: %v", err)
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating hash password: %v", err)
		return nil, err
	}

	user := model.User{
		Name:       req.Name,
		Password:   string(hash),
		Email:      req.Email,
		TOTPSecret: secret,
		Phone:      req.Phone,
		Role:       role,
		Status:     model.StatusActive,
	}

	createdUser, err := h.authService.Register(ctx, &user)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		return &auth.RegisterResponse{
			Status: h.createErrorResponse("Registration failed", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	loginResult, err := h.authService.Login(ctx, user.Email, req.Password, nil)
	if err != nil {
		log.Printf("Error generating tokens after registration: %v", err)
		return &auth.RegisterResponse{
			Status: h.createErrorResponse("Registration successful but token generation failed", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	rsp := &auth.RegisterResponse{
		User: h.modelUserToProtoUserInfo(loginResult.User),
		Tokens: &auth.TokenInfo{
			AccessToken:  loginResult.AccessToken,
			RefreshToken: loginResult.RefreshToken,
			ExpiresAt:    timestamppb.New(loginResult.ExpiresAt),
		},
		Status: &common.ResponseStatus{
			Success: true,
			Message: "User registered successfully",
		},
	}

	log.Printf("User registered sucessfully; id=%d", createdUser.ID)
	return rsp, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	log.Printf("Login called with: email=%s", req.Email)

	var expectedRole *model.UserRole
	if req.ExpectedRole != auth.UserRole_USER {
		modelRole := h.protoRoleToModel(req.ExpectedRole)
		expectedRole = &modelRole
	}

	loginResult, err := h.authService.Login(ctx, req.Email, req.Password, expectedRole)
	if err != nil {
		log.Printf("Error during login: %v", err)
		return &auth.LoginResponse{
			Status: h.createErrorResponse("Login failed", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	return &auth.LoginResponse{
		User: h.modelUserToProtoUserInfo(loginResult.User),
		Tokens: &auth.TokenInfo{
			AccessToken:  loginResult.AccessToken,
			RefreshToken: loginResult.RefreshToken,
			ExpiresAt:    timestamppb.New(loginResult.ExpiresAt),
		},
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Login successful",
		},
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	log.Printf("RefreshToken called with: token=%v", req.RefreshToken)

	rt, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("Error during refreshing token: %v", err)
		return &auth.RefreshTokenResponse{
			Status: h.createErrorResponse("Refresh token error", common.ErrorCode_INTERNAL_ERROR),
		}, nil
	}

	return &auth.RefreshTokenResponse{
		Tokens: &auth.TokenInfo{
			AccessToken:  rt.AccessToken,
			RefreshToken: rt.RefreshToken,
			ExpiresAt:    timestamppb.New(rt.ExpiresAt),
		},
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Refresh token succesful",
		},
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	log.Printf("Validated token called with: token=%s", req.Token)

	vr, err := h.authService.ValidateToken(ctx, req.Token, req.RequiredPermissions)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		return &auth.ValidateTokenResponse{
			Status: h.createErrorResponse("Validate token error", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	return &auth.ValidateTokenResponse{
		Valid: vr.Valid,
		User:  h.modelUserToProtoUserInfo(vr.User),
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Validate token successful",
		},
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	log.Printf("Logout called with: token=%s", req.Token)

	if err := h.authService.Logout(ctx, req.Token); err != nil {
		log.Printf("Error during logout: %v", err)
		return &auth.LogoutResponse{
			Status: h.createErrorResponse("Logout error", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	return &auth.LogoutResponse{
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Logout successful",
		},
	}, nil
}

func (h *AuthHandler) CreateAdmin(ctx context.Context, req *auth.CreateAdminRequest) (*auth.CreateAdminResponse, error) {
	log.Printf("Create admin called with: createdBy=%s", req.CreatedBy)

	secret, err := service.GenerateTOTPSecret(req.Name)
	if err != nil {
		log.Printf("Error generating TOTP: %v", err)
		return &auth.CreateAdminResponse{
			Status: h.createErrorResponse("Generate TOTP error", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating hash password: %v", err)
		return &auth.CreateAdminResponse{
			Status: h.createErrorResponse("Hash password error", common.ErrorCode_INTERNAL_ERROR),
		}, nil
	}

	role := h.protoRoleToModel(req.Role)

	admin := &model.User{
		Email:      req.Email,
		Password:   string(hash),
		TOTPSecret: secret,
		Name:       req.Name,
		Phone:      req.Phone,
		Role:       role,
		Status:     model.StatusActive,
	}

	newAdmin, err := h.authService.CreateAdmin(ctx, admin)
	if err != nil {
		log.Printf("Error creating admin: %v", err)
		return &auth.CreateAdminResponse{
			Status: h.createErrorResponse("Create admin error", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	return &auth.CreateAdminResponse{
		Admin: h.modelUserToProtoUserInfo(newAdmin),
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Create admin successful",
		},
	}, nil
}

func (h *AuthHandler) ListUsers(ctx context.Context, req *auth.ListUsersRequest) (*auth.ListUsersResponse, error) {
	log.Printf("ListUsers called with page=%d, limit=%d", req.Page, req.Limit)

	var roleFilter *model.UserRole
	var statusFilter *model.UserStatus

	if req.RoleFilter != auth.UserRole_USER {
		modelRole := h.protoRoleToModel(req.RoleFilter)
		roleFilter = &modelRole
	}

	if req.StatusFilter != auth.UserStatus_ACTIVE {
		modelStatus := h.protoStatusToModel(req.StatusFilter)
		statusFilter = &modelStatus
	}

	users, total, err := h.authService.ListUsers(ctx, int(req.Page), int(req.Limit), roleFilter, statusFilter)
	if err != nil {
		log.Printf("Error listing users: %v", err)
		return &auth.ListUsersResponse{
			Status: h.createErrorResponse("Failed to list users", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	protoUsers := make([]*auth.UserInfo, len(users))
	for i, user := range users {
		protoUsers[i] = h.modelUserToProtoUserInfo(user)
	}

	limit := int32(req.Limit)
	if limit == 0 {
		limit = 10
	}
	totalPages := (int32(total) + limit - 1) / limit
	currentPage := req.Page
	if currentPage == 0 {
		currentPage = 1
	}

	return &auth.ListUsersResponse{
		Users: protoUsers,
		Pagination: &common.PaginationInfo{
			Total:       int32(total),
			Page:        currentPage,
			Limit:       limit,
			TotalPages:  totalPages,
			HasNext:     currentPage < totalPages,
			HasPrevious: currentPage > 1,
		},
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Users retrieved successfully",
		},
	}, nil
}

func (h *AuthHandler) UpdateUserRole(ctx context.Context, req *auth.UpdateUserRoleRequest) (*auth.UpdateUserRoleResponse, error) {
	log.Printf("UpdateUserRole called with: user id=%s, updated by=%s", req.UserId, req.UpdatedBy)

	userId, err := strconv.Atoi(req.UserId)
	if err != nil {
		log.Printf("Error user id: %v", err)
		return &auth.UpdateUserRoleResponse{
			Status: h.createErrorResponse("Error user id", common.ErrorCode_INVALID_REQUEST),
		}, err
	}

	updatedBy, err := strconv.Atoi(req.UpdatedBy)
	if err != nil {
		log.Printf("Error updated by: %v", err)
		return &auth.UpdateUserRoleResponse{
			Status: h.createErrorResponse("Error updated by", common.ErrorCode_INVALID_REQUEST),
		}, err
	}

	if err := h.authService.UpdateUserRole(ctx, uint(userId), h.protoRoleToModel(req.NewRole), uint(updatedBy)); err != nil {
		log.Printf("Error update user role: %v", err)
		return &auth.UpdateUserRoleResponse{
			Status: h.createErrorResponse("Error update user role", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	return &auth.UpdateUserRoleResponse{
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Update user role successfully",
		},
	}, nil
}

func (h *AuthHandler) DeactivateUser(ctx context.Context, req *auth.DeactivateUserRequest) (*auth.DeactivateUserResponse, error) {
	log.Printf("DeactivateUser called with: user id=%s", req.UserId)

	userId, err := strconv.Atoi(req.UserId)
	if err != nil {
		log.Printf("Error user id: %v", err)
		return &auth.DeactivateUserResponse{
			Status: h.createErrorResponse("Error user id", common.ErrorCode_INVALID_REQUEST),
		}, err
	}

	updatedBy, err := strconv.Atoi(req.UpdatedBy)
	if err != nil {
		log.Printf("Error updated by: %v", err)
		return &auth.DeactivateUserResponse{
			Status: h.createErrorResponse("Error updated by", common.ErrorCode_INVALID_REQUEST),
		}, err
	}

	if err := h.authService.DeactivateUser(ctx, uint(userId), h.protoStatusToModel(req.NewStatus), req.Reason, uint(updatedBy)); err != nil {
		log.Printf("Error deactivate user: %v", err)
		return &auth.DeactivateUserResponse{
			Status: h.createErrorResponse("Error deactivate user", common.ErrorCode_INTERNAL_ERROR),
		}, err
	}

	return &auth.DeactivateUserResponse{
		Status: &common.ResponseStatus{
			Success: true,
			Message: "Deactivate user successfully",
		},
	}, nil
}

func (h *AuthHandler) modelUserToProtoUserInfo(user *model.User) *auth.UserInfo {
	userInfo := &auth.UserInfo{
		UserId:      strconv.Itoa(int(user.ID)),
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		Role:        h.modelRoleToProto(user.Role),
		Status:      h.modelStatusToProto(user.Status),
		Permissions: user.GetPermissions(),
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}

	if user.LastLogin != nil {
		userInfo.LastLogin = timestamppb.New(*user.LastLogin)
	}

	return userInfo
}

// Role conversion helpers
func (h *AuthHandler) protoRoleToModel(protoRole auth.UserRole) model.UserRole {
	switch protoRole {
	case auth.UserRole_ADMIN:
		return model.RoleAdmin
	case auth.UserRole_SUPER_ADMIN:
		return model.RoleSuperAdmin
	default:
		return model.RoleUser
	}
}

func (h *AuthHandler) modelRoleToProto(modelRole model.UserRole) auth.UserRole {
	switch modelRole {
	case model.RoleAdmin:
		return auth.UserRole_ADMIN
	case model.RoleSuperAdmin:
		return auth.UserRole_SUPER_ADMIN
	default:
		return auth.UserRole_USER
	}
}

// Status conversion helpers
func (h *AuthHandler) protoStatusToModel(protoStatus auth.UserStatus) model.UserStatus {
	switch protoStatus {
	case auth.UserStatus_INACTIVE:
		return model.StatusInactive
	case auth.UserStatus_SUSPENDED:
		return model.StatusSuspended
	default:
		return model.StatusActive
	}
}

func (h *AuthHandler) modelStatusToProto(modelStatus model.UserStatus) auth.UserStatus {
	switch modelStatus {
	case model.StatusInactive:
		return auth.UserStatus_INACTIVE
	case model.StatusSuspended:
		return auth.UserStatus_SUSPENDED
	default:
		return auth.UserStatus_ACTIVE
	}
}
