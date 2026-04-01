package repo

import (
	"context"
	"refresh-token/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	err := r.db.WithContext(ctx).Preload("Role").Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Role").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepo) DeleteUser(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *UserRepo) GetAllUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Preload("Role").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
