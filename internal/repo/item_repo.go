package repo

import (
	"context"
	"refresh-token/internal/model"

	"gorm.io/gorm"
)

type ItemRepo struct {
	db *gorm.DB
}

func NewItemRepo(db *gorm.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

func (r *ItemRepo) CreateItem(ctx context.Context, item *model.Item) (*model.Item, error) {
	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *ItemRepo) GetItemByID(ctx context.Context, id int) (*model.Item, error) {
	var item model.Item
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ItemRepo) UpdateItem(ctx context.Context, item *model.Item) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *ItemRepo) DeleteItem(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&model.Item{}, id).Error
}

func (r *ItemRepo) GetAllItems(ctx context.Context) ([]model.Item, error) {
	var items []model.Item
	err := r.db.WithContext(ctx).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}