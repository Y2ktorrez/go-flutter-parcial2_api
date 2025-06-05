package repositories

import "github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"

type ProjectRepository interface {
	Create(project *entity.Project) error
	FindByID(id string) (*entity.Project, error)
	FindAll() ([]entity.Project, error)
	Update(project *entity.Project) error
	Delete(id string) error
}
