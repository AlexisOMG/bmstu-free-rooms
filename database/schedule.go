package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/AlexisOMG/bmstu-free-rooms/service"
)

var (
	scheduleTable       = "schedule"
	schedulesFieldNames = []string{
		"audience_id",
		"lesson_id",
		"week_type",
		"week_day",
		"lesson_start",
		"lesson_end",
		"period",
	}
)

type schedule struct {
	ID         string     `db:"id"`
	AudienceID string     `db:"audience_id"`
	LessonID   string     `db:"lesson_id"`
	WeekType   string     `db:"week_type"`
	WeekDay    string     `db:"week_day"`
	Start      *time.Time `db:"lesson_start"`
	End        *time.Time `db:"lesson_end"`
	Period     int        `db:"period"`
}

func (s *schedule) toService() service.Schedule {
	return service.Schedule{
		ID:         s.ID,
		AudienceID: s.AudienceID,
		LessonID:   s.LessonID,
		WeekType:   s.WeekType,
		WeekDay:    s.WeekDay,
		Start:      s.Start,
		End:        s.End,
		Period:     s.Period,
	}
}

func (s *schedule) values() []interface{} {
	return []interface{}{
		s.ID,
		s.AudienceID,
		s.LessonID,
		s.WeekType,
		s.WeekDay,
		s.Start,
		s.End,
		s.Period,
	}
}

func scheduleToDB(s service.Schedule) schedule {
	return schedule{
		ID:         s.ID,
		AudienceID: s.AudienceID,
		LessonID:   s.LessonID,
		WeekType:   s.WeekType,
		WeekDay:    s.WeekDay,
		Start:      s.Start,
		End:        s.End,
		Period:     s.Period,
	}
}

func (d *Database) SaveSchedules(ctx context.Context, schedules ...service.Schedule) error {
	if len(schedules) == 0 {
		return nil
	}
	dbSchedules := make([]schedule, 0, len(schedules))
	for _, a := range schedules {
		dbSchedules = append(dbSchedules, scheduleToDB(a))
	}

	query := squirrel.Insert(scheduleTable).Columns(append([]string{"id"}, schedulesFieldNames...)...)

	for _, dbA := range dbSchedules {
		query = query.Values(dbA.values()...)
	}

	query = query.Suffix("ON CONFLICT DO NOTHING").PlaceholderFormat(squirrel.Dollar)

	sql, bound, err := query.ToSql()
	if err != nil {
		return err
	}

	if _, err = d.db.ExecContext(ctx, sql, bound...); err != nil {
		return fmt.Errorf("cannot insert query: %v, args %v, into %v: %w", sql, bound, scheduleTable, err)
	}

	return nil
}

func (d *Database) ListSchedules(ctx context.Context, filters *service.ScheduleFilters) ([]service.Schedule, error) {
	return nil, nil
}
