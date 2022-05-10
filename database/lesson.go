package database

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/AlexisOMG/bmstu-free-rooms/service"
)

var (
	lessonTable       = "lesson"
	lessonsFieldNames = []string{
		"name",
		"teacher_name",
		"kind",
	}
)

type lesson struct {
	ID          string  `db:"id"`
	Name        string  `db:"name"`
	TeacherName *string `db:"teacher_name"`
	Kind        *string `db:"kind"`
}

func (l *lesson) toService() service.Lesson {
	return service.Lesson{
		ID:          l.ID,
		Name:        l.Name,
		TeacherName: l.TeacherName,
		Kind:        l.Kind,
	}
}

func (l *lesson) values() []interface{} {
	return []interface{}{
		l.ID,
		l.Name,
		l.TeacherName,
		l.Kind,
	}
}

func lessonToDB(l service.Lesson) lesson {
	return lesson{
		ID:          l.ID,
		Name:        l.Name,
		TeacherName: l.TeacherName,
		Kind:        l.Kind,
	}
}

func (d *Database) SaveLessons(ctx context.Context, lessons ...service.Lesson) error {
	if len(lessons) == 0 {
		return nil
	}
	dbLessons := make([]lesson, 0, len(lessons))
	for _, a := range lessons {
		dbLessons = append(dbLessons, lessonToDB(a))
	}

	query := squirrel.Insert(lessonTable).Columns(append([]string{"id"}, lessonsFieldNames...)...)

	for _, dbA := range dbLessons {
		query = query.Values(dbA.values()...)
	}

	query = query.Suffix("ON CONFLICT DO NOTHING").PlaceholderFormat(squirrel.Dollar)

	sql, bound, err := query.ToSql()
	if err != nil {
		return err
	}

	if _, err = d.db.ExecContext(ctx, sql, bound...); err != nil {
		return fmt.Errorf("cannot insert query: %v, args %v, into %v: %w", query, bound, lessonTable, err)
	}

	return nil
}

func (d *Database) ListLessons(ctx context.Context, filters *service.LessonFilters) ([]service.Lesson, error) {
	return nil, nil
}
