package repositories

import (
	"errors"

	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"

	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepositoryImpl) FindByID(id string) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuario no encontrado")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindAll() ([]entity.User, error) {
	var users []entity.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *UserRepositoryImpl) Update(user *entity.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&entity.User{}, "id = ?", id).Error
}

func (r *UserRepositoryImpl) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
