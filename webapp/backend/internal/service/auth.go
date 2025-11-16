package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"backend/internal/repository"
	"backend/internal/service/utils"

	"go.opentelemetry.io/otel"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInternalServer  = errors.New("internal server error")
)

type AuthService struct {
	store *repository.Store
}

func NewAuthService(store *repository.Store) *AuthService {
	return &AuthService{store: store}
}

func (s *AuthService) Login(ctx context.Context, userName, password string) (string, time.Time, error) {
	ctx, span := otel.Tracer("service.auth").Start(ctx, "AuthService.Login")
	defer span.End()

	var sessionID string
	var expiresAt time.Time
	err := utils.WithTimeout(ctx, func(ctx context.Context) error {
		user, err := s.store.UserRepo.FindByUserName(ctx, userName)
		if err != nil {
			log.Printf("[Login] ユーザー検索失敗(userName: %s): %v", userName, err)
			if errors.Is(err, sql.ErrNoRows) {
				return ErrUserNotFound
			}
			return ErrInternalServer
		}

		// パスワード検証処理をトレース
		_, pwSpan := otel.Tracer("service.auth").Start(ctx, "password.verify")
		// SHA256でパスワードを直接ハッシュ化して比較（平文変数への保存を回避）
		hash := sha256.Sum256([]byte(password))
		hashedPassword := hex.EncodeToString(hash[:])
		pwSpan.End()

		if user.PasswordHash != hashedPassword {
			log.Printf("[Login] パスワード検証失敗")
			span.RecordError(ErrInvalidPassword)
			return ErrInvalidPassword
		}

		sessionDuration := 24 * time.Hour
		sessionID, expiresAt, err = s.store.SessionRepo.Create(ctx, user.UserID, sessionDuration)
		if err != nil {
			log.Printf("[Login] セッション生成失敗: %v", err)
			return ErrInternalServer
		}
		return nil
	})
	if err != nil {
		return "", time.Time{}, err
	}
	// log.Printf("Login successful for UserName '%s', session created.", userName)
	return sessionID, expiresAt, nil
}
