package service

import (
	"cleanArch_with_postgres/internal/entity"
	"cleanArch_with_postgres/internal/infrastructure/config"
	"cleanArch_with_postgres/internal/repository"
	"cleanArch_with_postgres/internal/viewmodel"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, vm viewmodel.RegisterRequest) (*viewmodel.RegisterResponse, error)
	Login(ctx context.Context, identifier, password string) (*viewmodel.LoginResponse, error)
	GetUserVMByUsername(ctx context.Context, paramUsername, tokenUsername string) (*viewmodel.UserVM, error)
	SearchUsers(ctx context.Context, prefix string, limit int) ([]viewmodel.UserVM, error)
	UpdateUser(ctx context.Context, username string, vm *viewmodel.UpdateRequest) (*viewmodel.UpdateResponse, error)
	DeleteUser(ctx context.Context, username string) error
}

type authService struct {
	ur repository.UserRepository
	br repository.BlogRepository
}

func NewAuthService(ur repository.UserRepository, br repository.BlogRepository) AuthService {
	return &authService{ur: ur, br: br}
}

func (s *authService) Register(ctx context.Context, vm viewmodel.RegisterRequest) (*viewmodel.RegisterResponse, error) {
	user := &entity.User{
		Email:    vm.Email,
		Username: vm.Username,
		Password: vm.Password,
		Role:     entity.UserRole(vm.Role),
	}

	if user == nil {
		return nil, errors.New("user is nil")
	}

	exist, err := s.ur.ExistUser(ctx, vm.Email, vm.Username)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New("email or username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(vm.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing error: %w", err)
	}
	user.Password = string(hashed)

	if user.Role == "" {
		return nil, errors.New("lütfen bir rol seçiniz: reader, writer, admin")
	}

	if user.Role == "writer" {
		user.Role = entity.RoleWriter
	}
	if user.Role == "reader" {
		user.Role = entity.RoleReader
	}

	if user.Role == "admin" { // *************ADMİN ONAYI İÇİN BİLDİRİM MEKANİZMASI YAP************* //
		user.Role = entity.RoleReader
	}

	user.Followers = []string{}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	resp := &viewmodel.RegisterResponse{
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return resp, s.ur.Create(ctx, user)
}

type accessToken struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	UserID   uint   `json:"user_id"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
}

func (s *authService) Login(ctx context.Context, identifier, password string) (*viewmodel.LoginResponse, error) {
	user, err := s.ur.GetByIdentifier(ctx, identifier)
	if user == nil || err != nil { // ***
		return nil, errors.New("user not found")
	}

	if user.DeletedAt.Valid {
		return nil, errors.New("user is deleted")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	// jwtSecret := os.Getenv("JWT_SECRET") // config'den çek ***
	jwtSecret := []byte(config.Get().Secret.JWTSecret)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessToken{
		Username: user.Username,
		UserID:   user.ID,
		Role:     string(user.Role),
		Exp:      time.Now().Add(time.Hour * 24).Unix(),
	}).SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	resp := &viewmodel.LoginResponse{
		Token:    token,
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return resp, nil
}

func (s *authService) GetUserVMByUsername(ctx context.Context, paramUsername, tokenUsername string) (*viewmodel.UserVM, error) {
	if paramUsername == "" {
		return nil, errors.New("invalid username")
	}

	tokenUser, err := s.ur.GetByUsername(ctx, paramUsername)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if tokenUser.DeletedAt.Valid {
		return nil, errors.New("The logged in user has been deleted")
	}

	user, err := s.ur.GetByUsername(ctx, paramUsername)
	if err != nil {
		return nil, err
	}

	return viewmodel.ToUserVM(user), nil
}

// implementasyon:
func (s *authService) SearchUsers(ctx context.Context, prefix string, limit int) ([]viewmodel.UserVM, error) {
	if len(prefix) == 0 {
		return []viewmodel.UserVM{}, nil
	}
	users, err := s.ur.SearchByUsernamePrefix(ctx, prefix, limit)
	if err != nil {
		return nil, err
	}
	out := make([]viewmodel.UserVM, 0, len(users))
	for _, u := range users {
		out = append(out, *viewmodel.ToUserVM(&u))
	}
	return out, nil
}

func (s *authService) UpdateUser(ctx context.Context, username string, vm *viewmodel.UpdateRequest) (*viewmodel.UpdateResponse, error) {
	if vm == nil {
		return nil, errors.New("user is nil")
	}

	if vm.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(vm.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		vm.Password = string(hashed)
	}
	user, err := s.ur.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	existUser, err := s.ur.ExistUser(ctx, vm.Email, vm.Username)
	if err != nil {
		return nil, err
	}
	if existUser {
		return nil, errors.New("email or username already exists")
	}

	oldUsername := user.Username

	if user.DeletedAt.Valid {
		return nil, errors.New("user is deleted")
	}
	if vm.Username != "" {
		user.Username = vm.Username
	}
	if vm.Email != "" {
		user.Email = vm.Email
	}
	if vm.Password != "" {
		user.Password = vm.Password
	}
	user.UpdatedAt = time.Now()

	if oldUsername != user.Username {
		err := s.br.UpdateAuthorUsername(ctx, oldUsername, user.Username)
		if err != nil {
			return nil, errors.New("failed to author username in blogs")
		}
	}

	resp := viewmodel.UpdateResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		UpdatedAt: user.UpdatedAt,
	}

	return &resp, s.ur.Update(ctx, username, user)
}

func (s *authService) DeleteUser(ctx context.Context, username string) error {
	if username == "" {
		return errors.New("Invalid username")
	}

	user, err := s.ur.GetByUsername(ctx, username)
	if err != nil {
		return errors.New("user not found")
	}

	if user.DeletedAt.Valid {
		return errors.New("user is deleted")
	}

	return s.ur.Delete(ctx, username)
}
