package repo

import (
	"context"
	"refresh-token/internal/model"

	"gorm.io/gorm"
)

type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) CreateSession(ctx context.Context, session *model.Session) (*model.Session,error) {
	err := r.db.Create(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepo) GetSessionByID(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	err := r.db.Where("session_id=?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) GetSessionByUserID(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	err := r.db.Where("user_id=?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) RevokeSession(ctx context.Context, id string) error {
	return r.db.Model(&model.Session{}).Where("session_id = ?", id).Update("is_revoked", true).Error
}

func (r *SessionRepo) RevokeAllSessionForUser(ctx context.Context, userID int) error {
	return r.db.Model(&model.Session{}).Where("user_id = ?", userID).Update("is_revoked", true).Error
}

func (r *SessionRepo) DeleteSession(ctx context.Context, id string) error {
	return r.db.Delete(&model.Session{}, "session_id = ?", id).Error
}