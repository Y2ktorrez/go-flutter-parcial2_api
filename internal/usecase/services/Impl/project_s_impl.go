package impl

import (
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/repositories"
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/usecase/services"
)

type ProjectServiceImpl struct {
	repo repositories.ProjectRepository
}

func NewProjectService(repo repositories.ProjectRepository) services.ProjectService {
	return &ProjectServiceImpl{repo: repo}
}

func (s *ProjectServiceImpl) CreateProject(project *entity.Project) error {
	return s.repo.Create(project)
}

func (s *ProjectServiceImpl) GetProjectByID(id string) (*entity.Project, error) {
	return s.repo.FindByID(id)
}

func (s *ProjectServiceImpl) GetAllProjects() ([]entity.Project, error) {
	return s.repo.FindAll()
}

func (s *ProjectServiceImpl) UpdateProject(project *entity.Project) error {
	return s.repo.Update(project)
}

func (s *ProjectServiceImpl) DeleteProject(id string) error {
	return s.repo.Delete(id)
}
