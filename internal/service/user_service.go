package service

import (
	"errors"
	"go-backend/internal/models"
	"go-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// 对密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.userRepo.CreateUser(user)
}

// ValidateUser 验证用户登录
func (s *UserService) ValidateUser(username, password string) (*models.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("用户名不存在")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}

// GetUserByID 根据用户ID获取用户信息
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

// ListUsers 获取用户列表（仅管理员可用）
func (s *UserService) ListUsers(current, size int, filters map[string]interface{}) ([]models.User, int64, error) {
	offset := (current - 1) * size
	return s.userRepo.List(offset, size, filters)
}

// IsAdmin 检查用户是否为管理员
func (s *UserService) IsAdmin(userID uint) (bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return false, err
	}
	return user.Role == models.RoleAdmin, nil
}
