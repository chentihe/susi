package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tihe/susi-auth-service/internal/model"
	"github.com/tihe/susi-auth-service/internal/repository"
	"github.com/tihe/susi-shared/events"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, user *model.User) (*model.User, error)
	Login(ctx context.Context, email, password string, expectedRole *model.UserRole) (*LoginResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error)
	ValidateToken(ctx context.Context, token string, requiredPermissions []string) (*ValidateResult, error)
	Logout(ctx context.Context, token string) error

	// Admin operations
	CreateAdmin(ctx context.Context, user *model.User) (*model.User, error)
	ListUsers(ctx context.Context, page, limit int, roleFilter *model.UserRole, statusFilter *model.UserStatus) ([]*model.User, int64, error)
	UpdateUserRole(ctx context.Context, userID uint, newRole model.UserRole, updatedBy uint) error
	DeactivateUser(ctx context.Context, userID uint, newStatus model.UserStatus, reason string, updatedBy uint) error
}

type LoginResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type ValidateResult struct {
	Valid       bool
	User        *model.User
	Permissions []string
	ExpiresAt   time.Time
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	kafkaProducer    *events.KafkaProducer
}

func NewAdminService(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, producer *events.KafkaProducer) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		kafkaProducer:    producer,
	}
}

func (s *authService) Register(ctx context.Context, user *model.User) (*model.User, error) {
	if err := s.validateUserInput(*user); err != nil {
		return nil, err
	}

	if user.Role == "" {
		user.Role = model.RoleUser
	}

	existingUser, _ := s.userRepo.GetByEmail(ctx, user.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Password = string(hashedPassword)
	user.Name = strings.TrimSpace(user.Name)
	user.Phone = strings.TrimSpace(user.Phone)
	user.Status = model.StatusActive

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string, expectedRole *model.UserRole) (*LoginResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive() {
		return nil, fmt.Errorf("account is %s", user.Status)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check role if specified (admin login endpoints)
	if expectedRole != nil && user.Role != *expectedRole {
		if *expectedRole == model.RoleAdmin && !user.IsAdmin() {
			return nil, errors.New("insufficient privileges")
		}
	}

	accessToken, err := GenerateJWT(user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshTokenModel := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenModel); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	now := time.Now()
	user.LastLogin = &now
	s.userRepo.Update(ctx, user)

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}, nil
}

func (s *authService) CreateAdmin(ctx context.Context, user *model.User) (*model.User, error) {
	creator, err := s.userRepo.GetByID(ctx, user.CreatedBy)
	if err != nil {
		return nil, errors.New("invalid creator")
	}

	if !creator.IsSuperAdmin() {
		return nil, errors.New("only super admins can create admin users")
	}

	if user.Role != model.RoleAdmin && user.Role != model.RoleSuperAdmin {
		return nil, errors.New("invalid admin role")
	}

	if user.Role == model.RoleSuperAdmin && !creator.IsSuperAdmin() {
		return nil, errors.New("insufficient privileges to create super admin")
	}

	admin, err := s.Register(ctx, user)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *authService) ListUsers(ctx context.Context, page, limit int, roleFilter *model.UserRole, statusFilter *model.UserStatus) ([]*model.User, int64, error) {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit
	// TODO: add role and status filter args
	return s.userRepo.List(ctx, offset, limit)
}

func (s *authService) UpdateUserRole(ctx context.Context, userID uint, newRole model.UserRole, updatedBy uint) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	updater, err := s.userRepo.GetByID(ctx, updatedBy)
	if err != nil {
		return err
	}

	if !updater.IsSuperAdmin() {
		return errors.New("only super admins can update user roles")
	}

	user.Role = newRole
	user.UpdatedBy = updatedBy

	return s.userRepo.Update(ctx, user)
}

func (s *authService) DeactivateUser(ctx context.Context, userID uint, newStatus model.UserStatus, reason string, updatedBy uint) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	updater, err := s.userRepo.GetByID(ctx, updatedBy)
	if err != nil {
		return err
	}

	if !updater.IsSuperAdmin() {
		return errors.New("insufficient privileges")
	}

	if user.IsSuperAdmin() && !updater.IsSuperAdmin() {
		return errors.New("cannot deactivate super admin")
	}

	user.Status = newStatus
	user.UpdatedBy = updatedBy

	return s.userRepo.Update(ctx, user)
}

func (s *authService) validateUserInput(user model.User) error {
	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(user.Password) == "" || len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if strings.TrimSpace(user.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	rt, err := s.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to generate refersh token")
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	rt.Token = newRefreshToken
	rt.ExpiresAt = expiresAt
	rt.UpdatedAt = time.Now()
	if err := s.refreshTokenRepo.Update(ctx, rt); err != nil {
		return nil, errors.New("failed to update refresh token")
	}

	accessToken, err := GenerateJWT(user.Name)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &TokenResult{
		AccessToken:  accessToken,
		RefreshToken: rt.Token,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string, requiredPermissions []string) (*ValidateResult, error) {
	if token == "" {
		return nil, errors.New("missing token")
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	user, err := s.userRepo.GetByName(ctx, claims.Username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &ValidateResult{
		Valid:       true,
		User:        user,
		Permissions: requiredPermissions,
		ExpiresAt:   claims.ExpiresAt.Time,
	}, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	return s.refreshTokenRepo.Delete(ctx, token)
}
