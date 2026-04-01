package repo

import (
	"context"
	"refresh-token/internal/model"

	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
}

func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) CreateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	err := r.db.WithContext(ctx).Create(role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepo) GetRoleByID(ctx context.Context, id int) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) GetAllRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepo) UpdateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *RoleRepo) DeleteRole(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}