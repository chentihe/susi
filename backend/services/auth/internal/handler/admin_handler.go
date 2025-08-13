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

func (h *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest, rsp *auth.RegisterResponse) error {
	log.Printf("Register called with: name=%s, email=%s", req.Name, req.Email)

	role := h.protoRoleToModel(req.Role)

	secret, err := service.GenerateTOTPSecret(req.Name)
	if err != nil {
		log.Printf("Error generating TOTP: %v", err)
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating hash password: %v", err)
		return err
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
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Registration failed",
			ErrorCode: "REGISTRATION_ERROR",
		}
		return err
	}

	loginResult, err := h.authService.Login(ctx, user.Email, req.Password, nil)
	if err != nil {
		log.Printf("Error generating tokens after registration: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Registration successful but token generation failed",
			ErrorCode: "TOKEN_ERROR",
		}
		return err
	}

	rsp.User = h.modelUserToProtoUserInfo(loginResult.User)
	rsp.Tokens = &auth.TokenInfo{
		AccessToken:  loginResult.AccessToken,
		RefreshToken: loginResult.RefreshToken,
		ExpiresAt:    timestamppb.New(loginResult.ExpiresAt),
	}
	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "User registered successfully",
	}

	log.Printf("User registered sucessfully; id=%d", createdUser.ID)
	return nil
}

func (h *AuthHandler) Login(ctx context.Context, req *auth.LoginRequest, rsp *auth.LoginResponse) error {
	log.Printf("Login called with: email=%s", req.Email)

	var expectedRole *model.UserRole
	if req.ExpectedRole != auth.UserRole_USER {
		modelRole := h.protoRoleToModel(req.ExpectedRole)
		expectedRole = &modelRole
	}

	loginResult, err := h.authService.Login(ctx, req.Email, req.Password, expectedRole)
	if err != nil {
		log.Printf("Error during login: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Login failed",
			ErrorCode: "LOGIN_ERROR",
		}
		return err
	}

	rsp.User = h.modelUserToProtoUserInfo(loginResult.User)
	rsp.Tokens = &auth.TokenInfo{
		AccessToken:  loginResult.AccessToken,
		RefreshToken: loginResult.RefreshToken,
		ExpiresAt:    timestamppb.New(loginResult.ExpiresAt),
	}
	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Login successful",
	}

	return nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest, rsp *auth.RefreshTokenResponse) error {
	log.Printf("RefreshToken called with: token=%v", req.RefreshToken)

	rt, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("Error during refreshing token: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Refresh token error",
			ErrorCode: "REFRESH_TOKEN_ERROR",
		}
		return nil
	}

	rsp.Tokens = &auth.TokenInfo{
		AccessToken:  rt.AccessToken,
		RefreshToken: rt.RefreshToken,
		ExpiresAt:    timestamppb.New(rt.ExpiresAt),
	}
	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Refresh token succesful",
	}

	return nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest, rsp *auth.ValidateTokenResponse) error {
	log.Printf("Validated token called with: token=%s", req.Token)

	vr, err := h.authService.ValidateToken(ctx, req.Token, req.RequiredPermissions)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Validate token error",
			ErrorCode: "VALIDATE_TOKEN_ERROR",
		}
		return err
	}

	rsp.Valid = vr.Valid
	rsp.User = h.modelUserToProtoUserInfo(vr.User)
	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Validate token successful",
	}

	return nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *auth.LogoutRequest, rsp *auth.LogoutResponse) error {
	log.Printf("Logout called with: token=%s", req.Token)

	if err := h.authService.Logout(ctx, req.Token); err != nil {
		log.Printf("Error during logout: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Logout error",
			ErrorCode: "LOGOUT_ERROR",
		}
		return err
	}

	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Logout successful",
	}

	return nil
}

func (h *AuthHandler) CreateAdmin(ctx context.Context, req *auth.CreateAdminRequest, rsp *auth.CreateAdminResponse) error {
	log.Printf("Create admin called with: createdBy=%s", req.CreatedBy)

	secret, err := service.GenerateTOTPSecret(req.Name)
	if err != nil {
		log.Printf("Error generating TOTP: %v", err)
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating hash password: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Hash pass error",
			ErrorCode: "HASH_PASS_ERROR",
		}
		return nil
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
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Create admin error",
			ErrorCode: "CREATE_ADMIN_ERROR",
		}
		return err
	}

	rsp.Admin = h.modelUserToProtoUserInfo(newAdmin)
	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Create admin successful",
	}

	return nil
}

func (h *AuthHandler) ListUsers(ctx context.Context, req *auth.ListUsersRequest, rsp *auth.ListUsersResponse) error {
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
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Failed to list users",
			ErrorCode: "LIST_ERROR",
		}
		return err
	}

	rsp.Users = make([]*auth.UserInfo, len(users))
	for i, user := range users {
		rsp.Users[i] = h.modelUserToProtoUserInfo(user)
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

	rsp.Pagination = &auth.PaginationInfo{
		Total:       int32(total),
		Page:        currentPage,
		Limit:       limit,
		TotalPages:  totalPages,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}

	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Users retrieved successfully",
	}

	return nil
}

func (h *AuthHandler) UpdateUserRole(ctx context.Context, req *auth.UpdateUserRoleRequest, rsp *auth.UpdateUserRoleResponse) error {
	log.Printf("UpdateUserRole called with: user id=%s, updated by=%s", req.UserId, req.UpdatedBy)

	userId, err := strconv.Atoi(req.UserId)
	if err != nil {
		log.Printf("Error user id: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Error user id",
			ErrorCode: "USER_ID_ERROR",
		}
		return err
	}

	updatedBy, err := strconv.Atoi(req.UpdatedBy)
	if err != nil {
		log.Printf("Error updated by: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Error updated by",
			ErrorCode: "UPDATED_BY_ERROR",
		}
		return err
	}

	if err := h.authService.UpdateUserRole(ctx, uint(userId), h.protoRoleToModel(req.NewRole), uint(updatedBy)); err != nil {
		log.Printf("Error update user role: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Error update user role",
			ErrorCode: "UPDATE_ROLE_ERROR",
		}
		return err
	}

	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Update user role successfully",
	}

	return nil
}

func (h *AuthHandler) DeactivateUser(ctx context.Context, req *auth.DeactivateUserRequest, rsp *auth.DeactivateUserResponse) error {
	log.Printf("DeactivateUser called with: user id=%s", req.UserId)

	userId, err := strconv.Atoi(req.UserId)
	if err != nil {
		log.Printf("Error user id: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Error user id",
			ErrorCode: "USER_ID_ERROR",
		}
		return err
	}

	updatedBy, err := strconv.Atoi(req.UpdatedBy)
	if err != nil {
		log.Printf("Error updated by: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Error updated by",
			ErrorCode: "UPDATED_BY_ERROR",
		}
		return err
	}

	if err := h.authService.DeactivateUser(ctx, uint(userId), h.protoStatusToModel(req.NewStatus), req.Reason, uint(updatedBy)); err != nil {
		log.Printf("Error deactivate user: %v", err)
		rsp.Status = &auth.ResponseStatus{
			Success:   false,
			Message:   "Error deactivate user",
			ErrorCode: "DEACTIVATE_ERROR",
		}
		return err
	}

	rsp.Status = &auth.ResponseStatus{
		Success: true,
		Message: "Deactivate user successfully",
	}

	return nil
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
