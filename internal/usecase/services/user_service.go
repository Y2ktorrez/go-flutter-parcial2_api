package services

import "github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"

type UserService interface {
	CreateUser(user *entity.User) error
	GetUserByID(id string) (*entity.User, error)
	GetAllUsers() ([]entity.User, error)
	UpdateUser(user *entity.User) error
	DeleteUser(id string) error
}
