package repo

import (
	"context"
	"encoding/json"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"refresh-token/internal/infra/redis"
	"refresh-token/internal/model"
	"time"
)

type SessionRepo struct {
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{}
}

func (r *SessionRepo) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	session.CreatedAt = time.Now()
	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return session, nil
	}

	sessionKey := fmt.Sprintf("session:%s", session.SessionID)
	userSetKey := fmt.Sprintf("user_sessions:%d", session.UserID)

	// Execute commands in a transaction (MULTI/EXEC)
	_, err = redis.RedisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.Set(ctx, sessionKey, data, ttl)
		pipe.SAdd(ctx, userSetKey, session.SessionID)
		pipe.Expire(ctx, userSetKey, 7*24*time.Hour)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to save session in redis transaction: %w", err)
	}

	return session, nil
}

func (r *SessionRepo) GetSessionByID(ctx context.Context, id string) (*model.Session, error) {
	key := fmt.Sprintf("session:%s", id)
	val, err := redis.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var session model.Session
	err = json.Unmarshal([]byte(val), &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) RevokeSession(ctx context.Context, id string) error {
	session, err := r.GetSessionByID(ctx, id)
	if err != nil {
		if err == goredis.Nil {
			return nil
		}
		return err
	}

	sessionKey := fmt.Sprintf("session:%s", id)
	userSetKey := fmt.Sprintf("user_sessions:%d", session.UserID)

	// Execute deletion in a transaction
	_, err = redis.RedisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.Del(ctx, sessionKey)
		pipe.SRem(ctx, userSetKey, id)
		return nil
	})

	return err
}

func (r *SessionRepo) RevokeAllSessionForUser(ctx context.Context, userID int) error {
	userSetKey := fmt.Sprintf("user_sessions:%d", userID)
	sessionIDs, err := redis.RedisClient.SMembers(ctx, userSetKey).Result()
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// Delete each individual session and the tracking set in a transaction
	_, err = redis.RedisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		for _, id := range sessionIDs {
			pipe.Del(ctx, fmt.Sprintf("session:%s", id))
		}
		pipe.Del(ctx, userSetKey)
		return nil
	})

	return err
}

func (r *SessionRepo) DeleteSession(ctx context.Context, id string) error {
	return r.RevokeSession(ctx, id)
}