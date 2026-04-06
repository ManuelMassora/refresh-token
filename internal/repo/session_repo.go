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
        pipe.Expire(ctx, userSetKey, 30*24*time.Hour) // renova o TTL a cada nova sessão
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

func (r *SessionRepo) ReplaceSession(ctx context.Context, oldSessionID string, newSession *model.Session) error {
    newSession.CreatedAt = time.Now()

    ttl := time.Until(newSession.ExpiresAt)
    if ttl <= 0 {
        return fmt.Errorf("session already expired")
    }

    data, err := json.Marshal(newSession)
    if err != nil {
        return err
    }

    oldSessionKey := fmt.Sprintf("session:%s", oldSessionID)
    newSessionKey := fmt.Sprintf("session:%s", newSession.SessionID)
    userSetKey := fmt.Sprintf("user_sessions:%s", newSession.UserID)

    _, err = redis.RedisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
        // Delete old session
        pipe.Del(ctx, oldSessionKey)
        pipe.SRem(ctx, userSetKey, oldSessionID)

        // Create new session
        pipe.Set(ctx, newSessionKey, data, ttl)
        pipe.SAdd(ctx, userSetKey, newSession.SessionID)
        pipe.Expire(ctx, userSetKey, 30*24*time.Hour) // renova o TTL

        // Update device map if applicable
        if newSession.DeviceID != "" {
            pipe.Set(ctx, "device:session:"+newSession.DeviceID, newSession.SessionID, 7*24*time.Hour)
        }
        return nil
    })
    if err != nil {
        return fmt.Errorf("failed to replace session: %w", err)
    }
    return nil
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
func (r *SessionRepo) GetActiveSessionsForUser(ctx context.Context, userID string) ([]*model.Session, error) {
	userSetKey := fmt.Sprintf("user_sessions:%s", userID)
	ids, err := redis.RedisClient.SMembers(ctx, userSetKey).Result()
	if err != nil {
		return nil, err
	}

	var active []*model.Session
	for _, id := range ids {
		session, err := r.GetSessionByID(ctx, id)
		if err != nil {
			if errors.Is(err, goredis.Nil) {
				// Lazy cleanup: a sessão expirou no Redis mas ainda está no set
				redis.RedisClient.SRem(ctx, userSetKey, id)
				continue
			}
			return nil, err
		}
		active = append(active, session)
	}
	return active, nil
}
