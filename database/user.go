package database

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/AlexisOMG/bmstu-free-rooms/service"
)

var (
	userTable       = "user_info"
	usersFieldNames = []string{
		"telegram_id",
		"username",
		"phone",
	}
)

type user struct {
	ID         string  `db:"id"`
	TelegramID string  `db:"telegram_id"`
	Username   *string `db:"username"`
	Phone      *string `db:"phone"`
}

func (u *user) toService() service.User {
	return service.User{
		ID:         u.ID,
		TelegramID: u.TelegramID,
		Username:   u.Username,
		Phone:      u.Phone,
	}
}

func (u *user) values() []interface{} {
	return []interface{}{
		u.ID,
		u.TelegramID,
		u.Username,
		u.Phone,
	}
}

func userToDB(u *service.User) user {
	return user{
		ID:         u.ID,
		TelegramID: u.TelegramID,
		Username:   u.Username,
		Phone:      u.Phone,
	}
}

func usersToService(users []user) []service.User {
	res := make([]service.User, 0, len(users))
	for _, usr := range users {
		res = append(res, usr.toService())
	}
	return res
}

func (d *Database) SaveUser(ctx context.Context, user *service.User) error {
	dbUser := userToDB(user)

	query := squirrel.
		Insert(userTable).
		Columns(append([]string{"id"}, usersFieldNames...)...).
		Values(dbUser.values()...)

	query = query.Suffix("ON CONFLICT ON CONSTRAINT telegram_unique DO NOTHING").PlaceholderFormat(squirrel.Dollar)

	sql, bound, err := query.ToSql()
	if err != nil {
		return err
	}

	if _, err = d.db.ExecContext(ctx, sql, bound...); err != nil {
		return fmt.Errorf("cannot insert query: %v, args %v, into %v: %w", query, bound, userTable, err)
	}

	return nil
}

func (d *Database) ListUsers(ctx context.Context, filters *service.UserFilters) ([]service.User, error) {
	res := []user{}

	query := squirrel.Select(append([]string{"id"}, usersFieldNames...)...).
		From(userTable).PlaceholderFormat(squirrel.Dollar)

	if len(filters.TelegramIDs) > 0 {
		query = query.Where(squirrel.Eq{"telegram_id": filters.TelegramIDs})
	}

	sqlText, bound, err := query.ToSql()
	if err != nil {
		return []service.User{}, err
	}

	if err = d.db.SelectContext(ctx, &res, sqlText, bound...); err != nil {
		return []service.User{}, mapErrors(err, "cannot select "+userTable+": %w")
	}

	return usersToService(res), nil
}
