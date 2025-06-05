package services

import (
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/repositories"
)

type UserServiceImpl struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) CreateUser(user *entity.User) error {
	return s.repo.Create(user)
}

func (s *UserServiceImpl) GetUserByID(id string) (*entity.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserServiceImpl) GetAllUsers() ([]entity.User, error) {
	return s.repo.FindAll()
}

func (s *UserServiceImpl) UpdateUser(user *entity.User) error {
	return s.repo.Update(user)
}

func (s *UserServiceImpl) DeleteUser(id string) error {
	return s.repo.Delete(id)
}
