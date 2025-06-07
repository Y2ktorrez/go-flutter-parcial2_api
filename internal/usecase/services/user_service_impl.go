package services

import (
	"errors"
	"time"

	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/dto"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/repositories"
	"github.com/golang-jwt/jwt/v5"
)

type UserServiceImpl struct {
	repo      repositories.UserRepository
	jwtSecret string
}

func NewUserService(repo repositories.UserRepository, jwtS string) UserService {
	return &UserServiceImpl{
		repo:      repo,
		jwtSecret: jwtS,
	}
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

func (s *UserServiceImpl) Login(request *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Buscar usuario por email
	user, err := s.repo.FindByEmail(request.Email)
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	// Verificar contraseña
	if !user.CheckPassword(request.Password) {
		return nil, errors.New("credenciales inválidas")
	}

	// Generar JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, errors.New("error al generar token")
	}

	// Preparar respuesta
	response := &dto.LoginResponse{
		Token: token,
	}
	response.User.ID = user.ID.String()
	response.User.Email = user.Email
	response.User.Name = user.Name

	return response, nil
}

func (s *UserServiceImpl) Signup(request *dto.SignupRequest) (*dto.SignupResponse, error) {
	// Verificar si el email ya existe
	exists, err := s.repo.EmailExists(request.Email)
	if err != nil {
		return nil, errors.New("error al verificar email")
	}
	if exists {
		return nil, errors.New("el email ya está registrado")
	}

	// Crear nuevo usuario
	user := &entity.User{
		Email:    request.Email,
		Password: request.Password,
		Name:     request.Name,
	}

	// Encriptar contraseña
	if err := user.HashPassword(); err != nil {
		return nil, errors.New("error al procesar contraseña")
	}

	// Guardar en base de datos
	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("error al crear usuario")
	}

	// Generar JWT token automáticamente después del registro
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, errors.New("usuario creado pero error al generar token")
	}

	// Preparar respuesta con token incluido
	response := &dto.SignupResponse{
		Message: "Usuario creado exitosamente",
		Token:   token, // ← Token incluido automáticamente
	}
	response.User.ID = user.ID.String()
	response.User.Email = user.Email
	response.User.Name = user.Name

	return response, nil
}

func (s *UserServiceImpl) generateJWT(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Expira en 7 días
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
