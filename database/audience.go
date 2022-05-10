package database

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/AlexisOMG/bmstu-free-rooms/service"
)

var (
	audienceTable       = "audience"
	audiencesFieldNames = []string{
		"number",
		"building",
		"floor",
		"suffix",
	}
)

type audience struct {
	ID       string  `db:"id"`
	Number   string  `db:"number"`
	Building string  `db:"building"`
	Floor    int     `db:"floor"`
	Suffix   *string `db:"suffix"`
}

func (a *audience) toService() service.Audience {
	return service.Audience{
		ID:       a.ID,
		Number:   a.Number,
		Building: a.Building,
		Floor:    a.Floor,
		Suffix:   a.Suffix,
	}
}

func (a *audience) values() []interface{} {
	return []interface{}{
		a.ID,
		a.Number,
		a.Building,
		a.Floor,
		a.Suffix,
	}
}

func audienceToDB(a service.Audience) audience {
	return audience{
		ID:       a.ID,
		Number:   a.Number,
		Building: a.Building,
		Floor:    a.Floor,
		Suffix:   a.Suffix,
	}
}

func (d *Database) SaveAudiences(ctx context.Context, audiences ...service.Audience) error {
	if len(audiences) == 0 {
		return nil
	}
	dbAudiences := make([]audience, 0, len(audiences))
	for _, a := range audiences {
		dbAudiences = append(dbAudiences, audienceToDB(a))
	}

	query := squirrel.Insert(audienceTable).Columns(append([]string{"id"}, audiencesFieldNames...)...)

	for _, dbA := range dbAudiences {
		query = query.Values(dbA.values()...)
	}

	query = query.Suffix("ON CONFLICT ON CONSTRAINT audience_unique DO NOTHING").PlaceholderFormat(squirrel.Dollar)

	sql, bound, err := query.ToSql()
	if err != nil {
		return err
	}

	if _, err = d.db.ExecContext(ctx, sql, bound...); err != nil {
		return fmt.Errorf("cannot insert query: %v, args %v, into %v: %w", query, bound, audienceTable, err)
	}

	return nil
}

func (d *Database) ListAudienceByNumber(ctx context.Context, number string, suffix *string) (service.Audience, error) {
	res := service.Audience{}
	query := squirrel.Select(append([]string{"id"}, audiencesFieldNames...)...).
		From(audienceTable).PlaceholderFormat(squirrel.Dollar)
	query = query.Where(squirrel.Eq{"number": number})
	if suffix != nil {
		query = query.Where(squirrel.Eq{"suffix": suffix})
	}

	sqlText, bound, err := query.ToSql()

	if err != nil {
		return service.Audience{}, fmt.Errorf("failed to build selection %v SQL: %w", audienceTable, err)
	}

	if err = d.db.GetContext(ctx, &res, sqlText, bound...); err != nil {
		return service.Audience{}, mapErrors(err, "cannot select "+audienceTable+": %w")
	}

	return res, nil
}

func (s *Database) ListEmptyAudiences(ctx context.Context, filters *service.EmptyAudiencesFilter) ([]service.Audience, error) {
	return nil, nil
}
