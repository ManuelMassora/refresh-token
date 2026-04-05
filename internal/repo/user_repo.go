package repo

import (
	"context"
	"fmt"
	"refresh-token/internal/infra/redis"
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

func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Role").Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateUserName(ctx context.Context, id string, username string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("username", username).Error
}

func (r *UserRepo) UpdateUserPassword(ctx context.Context, id string, password string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("password", password).Error
}

func (r *UserRepo) UpdateUserRole(ctx context.Context, id string, roleID int64) error {
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("role_id", roleID).Error
	if err == nil {
		key := fmt.Sprintf("user:role:%s", id)
		redis.RedisClient.Del(ctx, key)
	}
	return err
}

func (r *UserRepo) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}

func (r *UserRepo) GetAllUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Preload("Role").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
