package repositories

import "github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"

type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id string) (*entity.User, error)
	FindAll() ([]entity.User, error)
	Update(user *entity.User) error
	Delete(id string) error
}
