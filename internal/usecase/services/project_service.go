package services

import "github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"

type ProjectService interface {
	CreateProject(project *entity.Project) error
	GetProjectByID(id string) (*entity.Project, error)
	GetAllProjects() ([]entity.Project, error)
	UpdateProject(project *entity.Project) error
	DeleteProject(id string) error
}
