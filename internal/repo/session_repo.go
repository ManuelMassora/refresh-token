package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"refresh-token/internal/infra/redis"
	"refresh-token/internal/model"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type SessionRepo struct {
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{}
}

func (r *SessionRepo) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
    session.CreatedAt = time.Now()

    ttl := time.Until(session.ExpiresAt)
    if ttl <= 0 {
        return nil, fmt.Errorf("session already expired")
    }

    data, err := json.Marshal(session)
    if err != nil {
        return nil, err
    }

    sessionKey := fmt.Sprintf("session:%s", session.SessionID)
    userSetKey := fmt.Sprintf("user_sessions:%s", session.UserID)

    _, err = redis.RedisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
        pipe.Set(ctx, sessionKey, data, ttl)
        pipe.SAdd(ctx, userSetKey, session.SessionID)
        pipe.ExpireNX(ctx, userSetKey, 30*24*time.Hour) // só define se não existir
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("failed to save session: %w", err)
    }
    return session, nil
}

func (r *SessionRepo) DeleteSession(ctx context.Context, id string) error {
    session, err := r.GetSessionByID(ctx, id)
    if err != nil {
        if errors.Is(err, goredis.Nil) {
            return nil
        }
        return err
    }

    sessionKey := fmt.Sprintf("session:%s", id)
    userSetKey := fmt.Sprintf("user_sessions:%s", session.UserID)

    _, err = redis.RedisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
        pipe.Del(ctx, sessionKey)
        pipe.SRem(ctx, userSetKey, id)
        return nil
    })
    return err
}

func (r *SessionRepo) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
    userSetKey := fmt.Sprintf("user_sessions:%s", userID)

    sessionIDs, err := redis.RedisClient.SMembers(ctx, userSetKey).Result()
    if err != nil {
        return err
    }
    if len(sessionIDs) == 0 {
        return nil
    }

    _, err = redis.RedisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
        for _, id := range sessionIDs {
            pipe.Del(ctx, fmt.Sprintf("session:%s", id))
        }
        pipe.Del(ctx, userSetKey)
        return nil
    })
    return err
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