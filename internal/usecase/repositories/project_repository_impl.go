package repositories

import (
	"github.com/Y2ktorrez/go-flutter-parcial2_api/internal/entity"

	"gorm.io/gorm"
)

type ProjectRepositoryImpl struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &ProjectRepositoryImpl{db: db}
}

func (r *ProjectRepositoryImpl) Create(project *entity.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepositoryImpl) FindByID(id string) (*entity.Project, error) {
	var project entity.Project
	err := r.db.First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepositoryImpl) FindAll() ([]entity.Project, error) {
	var projects []entity.Project
	err := r.db.Find(&projects).Error
	return projects, err
}

func (r *ProjectRepositoryImpl) Update(project *entity.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&entity.Project{}, id).Error
}
