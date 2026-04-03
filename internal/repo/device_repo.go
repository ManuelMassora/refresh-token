package repo

import (
	"context"
	"refresh-token/internal/model"
	"time"

	"gorm.io/gorm"
)

type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) CreateDevice(ctx context.Context, device *model.Device) (*model.Device, error) {
	err := r.db.WithContext(ctx).Create(device).Error
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (r *DeviceRepo) GetDeviceByID(ctx context.Context, id string) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).First(&device, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepo) GetDeviceByFingerprint(ctx context.Context, fingerprint string) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).First(&device, "fingerprint = ?", fingerprint).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepo) UpdateDevice(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *DeviceRepo) UpdateLastSeen(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", id).Update("last_seen", time.Now()).Error
}

func (r *DeviceRepo) DeleteDevice(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Device{}, "id = ?", id).Error
}

func (r *DeviceRepo) GetAllDevices(ctx context.Context) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.WithContext(ctx).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}