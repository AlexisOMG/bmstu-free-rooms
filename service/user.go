package service

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	ID         string
	TelegramID string
	Username   *string
	Phone      *string
}

type UserFilters struct {
	TelegramIDs []string
}

func (s *Service) SaveUser(ctx context.Context, user *User) (string, error) {
	user.ID = uuid.NewString()
	if user.TelegramID == "" {
		return "", &ValidationError{
			ObjectKind: "User",
			Message:    "empty telegram ID",
		}
	}
	err := s.scheduleStorage.SaveUser(ctx, user)
	return user.ID, err
}

func (s *Service) ListUsers(ctx context.Context, filters *UserFilters) ([]User, error) {
	return s.scheduleStorage.ListUsers(ctx, filters)
}
