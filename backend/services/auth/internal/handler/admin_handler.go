package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"strings"
	"time"

	"github.com/tihe/susi-auth-service/internal/model"
	"github.com/tihe/susi-auth-service/internal/service"
	"github.com/tihe/susi-proto/auth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServiceHandler struct {
	authService service.AuthService
}

func NewAuthServiceHandler(authService service.AuthService) *AuthServiceHandler {
	return &AuthServiceHandler{
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

func (h *AuthServiceHandler) Register(ctx context.Context, req *auth.RegisterRequest, rsp *auth.RegisterResponse) error {
	log.Printf("Register called with: name=%s, email=%s", req.Name, req.Email)

	secret, err := service.GenerateTOTPSecret(req.Name)
	if err != nil {
		log.Printf("Error generating TOTP: %v", err)
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating hash password: %v", err)
	}

	user := model.User{
		Name:       req.Name,
		Password:   string(hash),
		Email:      req.Email,
		TOTPSecret: secret,
		Phone:      req.Phone,
		Role:       model.RoleUser,
		Status:     model.StatusActive,
	}

	createdUser, err := h.authService.Register(ctx, &user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}

	rsp. = h.modelToProto(createdUser)
	// TODO: need to refactor proto
	rsp.Message = "User created successfully"

	log.Printf("User created successfully: id=%d", admin.ID)
	return nil
}

func (h *AuthServiceHandler) Login(ctx context.Context, in *LoginRequest, out *LoginResponse) error {
	return h.AuthServiceHandler.Login(ctx, in, out)
}

func (h *AuthServiceHandler) RefreshToken(ctx context.Context, in *RefreshTokenRequest, out *RefreshTokenResponse) error {
	return h.AuthServiceHandler.RefreshToken(ctx, in, out)
}

func (h *AuthServiceHandler) ValidateToken(ctx context.Context, in *ValidateTokenRequest, out *ValidateTokenResponse) error {
	return h.AuthServiceHandler.ValidateToken(ctx, in, out)
}

func (h *AuthServiceHandler) Logout(ctx context.Context, in *LogoutRequest, out *LogoutResponse) error {
	return h.AuthServiceHandler.Logout(ctx, in, out)
}

func (h *AuthServiceHandler) CreateAdmin(ctx context.Context, in *CreateAdminRequest, out *CreateAdminResponse) error {
	return h.AuthServiceHandler.CreateAdmin(ctx, in, out)
}

func (h *AuthServiceHandler) ListUsers(ctx context.Context, in *ListUsersRequest, out *ListUsersResponse) error {
	return h.AuthServiceHandler.ListUsers(ctx, in, out)
}

func (h *AuthServiceHandler) UpdateUserRole(ctx context.Context, in *UpdateUserRoleRequest, out *UpdateUserRoleResponse) error {
	return h.AuthServiceHandler.UpdateUserRole(ctx, in, out)
}

func (h *AuthServiceHandler) DeactivateUser(ctx context.Context, in *DeactivateUserRequest, out *DeactivateUserResponse) error {
	return h.AuthServiceHandler.DeactivateUser(ctx, in, out)
}

// -------------------------------------------------------------------------------------------------------------------------

func (h *AdminHandler) CreateAdmin(ctx context.Context, req *admin.CreateAdminRequest, rsp *admin.CreateAdminResponse) error {
	log.Printf("CreateAdmin called with: name=%s, email=%s", req.Name, req.Email)

	secret, err := service.GenerateTOTPSecret(req.Name)
	if err != nil {
		log.Printf("Error generating TOTP: %v", err)
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating hash password: %v", err)
	}

	admin := model.Admin{
		Name:         req.Name,
		PasswordHash: string(hash),
		Email:        req.Email,
		TOTPSecret:   secret,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdAdmin, err := h.adminService.CreateAdmin(ctx, &admin)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}

	rsp.Admin = h.modelToProto(createdAdmin)
	rsp.Message = "User created successfully"

	log.Printf("User created successfully: id=%d", admin.ID)
	return nil
}

func (h *AdminHandler) GetAdmin(ctx context.Context, req *admin.GetAdminRequest, rsp *admin.GetAdminResponse) error {
	log.Printf("GetAdmin called with id=%d", req.Id)

	adminModel, err := h.adminService.GetAdminByID(ctx, int(req.Id))
	if err != nil {
		log.Printf("Error getting admin: %v", err)
		return err
	}

	rsp.Admin = h.modelToProto(adminModel)
	return nil
}

func (h *AdminHandler) UpdateAdmin(ctx context.Context, req *admin.UpdateAdminRequest, rsp *admin.UpdateAdminResponse) error {
	log.Printf("UpdateAdmin called with id=%d", req.Id)

	admin := model.Admin{
		Name:      req.Name,
		Email:     req.Email,
		UpdatedAt: time.Now(),
	}

	updatedAdmin, err := h.adminService.UpdateAdmin(ctx, &admin)
	if err != nil {
		log.Printf("Error updating admin: %v", err)
		return err
	}

	rsp.Admin = h.modelToProto(updatedAdmin)
	rsp.Message = "Admin updated successfully"
	return nil
}

func (h *AdminHandler) DeleteAdmin(ctx context.Context, req *admin.DeleteAdminRequest, rsp *admin.DeleteAdminResponse) error {
	log.Panicf("DeleteAdmin called with id=%d", req.Id)

	err := h.adminService.DeleteAdmin(ctx, int(req.Id))
	if err != nil {
		log.Panicf("Error deleting admin: %v", err)
		return err
	}

	rsp.Message = "Admin deleted successfully"
	return nil
}

func (h *AdminHandler) ListAdmins(ctx context.Context, req *admin.ListAdminsRequest, rsp *admin.ListAdminsResponse) error {
	log.Printf("ListAdmins called with page=%d limit=%d", req.Page, req.Limit)

	admins, total, err := h.adminService.ListAdmins(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		log.Printf("Error listing admins: %v", err)
		return err
	}

	rsp.Admins = make([]*admin.Admin, len(admins))
	for i, adminModel := range admins {
		rsp.Admins[i] = h.modelToProto(adminModel)
	}

	rsp.Total = int32(total)
	rsp.Page = req.Page
	rsp.Limit = req.Limit

	return nil
}

func (h *AuthServiceHandler) modelToProto(u *model.User) *auth.UserInfo {
	return &auth.UserInfo{
		UserId:    int32(u.ID),
		Email:     u.Email,
		Name:      u.Name,
		Phone:     u.Phone,
		Role:      auth.UserRole(auth.UserRole_value[strings.ToUpper(string(u.Role))]),
		Status:    auth.UserStatus(auth.UserStatus_value[strings.ToUpper(string(u.Status))]),
		CreatedAt: timestamppb.New(u.CreatedAt),
		LastLogin: timestamppb.New(*u.LastLogin),
	}
}
