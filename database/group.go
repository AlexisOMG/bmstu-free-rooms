package database

import (
	"context"
	"fmt"
	"ics/service"

	"github.com/Masterminds/squirrel"
)

var (
	groupTable       = "groups"
	groupsFieldNames = []string{
		"name",
	}
	groupLessonTable       = "group_lesson"
	groupLessonsFieldNames = []string{
		"group_id",
		"lesson_id",
	}
)

type group struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func (l *group) toService() service.Group {
	return service.Group{
		ID:   l.ID,
		Name: l.Name,
	}
}

func (l *group) values() []interface{} {
	return []interface{}{
		l.ID,
		l.Name,
	}
}

func groupToDB(l service.Group) group {
	return group{
		ID:   l.ID,
		Name: l.Name,
	}
}

func groupsToService(gs []group) []service.Group {
	res := make([]service.Group, 0, len(gs))
	for i := range gs {
		res = append(res, gs[i].toService())
	}
	return res
}

func (d *Database) SaveGroups(ctx context.Context, groups ...service.Group) error {
	if len(groups) == 0 {
		return nil
	}
	dbGroups := make([]group, 0, len(groups))
	for _, a := range groups {
		dbGroups = append(dbGroups, groupToDB(a))
	}

	query := squirrel.Insert(groupTable).Columns(append([]string{"id"}, groupsFieldNames...)...)

	for _, dbA := range dbGroups {
		query = query.Values(dbA.values()...)
	}

	query = query.Suffix("ON CONFLICT DO NOTHING").PlaceholderFormat(squirrel.Dollar)

	sql, bound, err := query.ToSql()
	if err != nil {
		return err
	}

	if _, err = d.db.ExecContext(ctx, sql, bound...); err != nil {
		return fmt.Errorf("cannot insert query: %v, args %v, into %v: %w", query, bound, groupTable, err)
	}

	return nil
}

func (d *Database) ListGroups(ctx context.Context, filters *service.GroupFilters) ([]service.Group, error) {
	res := make([]group, 0)
	query := squirrel.Select(append([]string{"id"}, groupsFieldNames...)...).
		From(groupTable).PlaceholderFormat(squirrel.Dollar)
	if filters.Name != nil {
		query = query.Where(squirrel.Eq{"name": filters.Name})
	}

	sqlText, bound, err := query.ToSql()
	if err != nil {
		return []service.Group{}, fmt.Errorf("failed to build selection %v SQL: %w", groupTable, err)
	}

	if err := d.db.SelectContext(ctx, &res, sqlText, bound...); err != nil {
		return []service.Group{}, mapErrors(err, "cannot select "+groupTable+": %w")
	}

	if len(res) == 0 {
		return []service.Group{}, service.ErrorNotFound
	}

	return groupsToService(res), nil
}

type groupLesson struct {
	ID       string `db:"id"`
	GroupID  string `db:"group_id"`
	LessonID string `db:"lesson_id"`
}

func (gl *groupLesson) toService() service.GroupLesson {
	return service.GroupLesson{
		ID:       gl.ID,
		GroupID:  gl.GroupID,
		LessonID: gl.LessonID,
	}
}

func (gl *groupLesson) values() []interface{} {
	return []interface{}{
		gl.ID,
		gl.GroupID,
		gl.LessonID,
	}
}

func groupLessonToDB(gl service.GroupLesson) groupLesson {
	return groupLesson{
		ID:       gl.ID,
		GroupID:  gl.GroupID,
		LessonID: gl.LessonID,
	}
}

func (d *Database) SaveGroupLessons(ctx context.Context, groupLessons ...service.GroupLesson) error {
	if len(groupLessons) == 0 {
		return nil
	}
	dbGroupLessons := make([]groupLesson, 0, len(groupLessons))
	for _, a := range groupLessons {
		dbGroupLessons = append(dbGroupLessons, groupLessonToDB(a))
	}

	query := squirrel.Insert(groupLessonTable).Columns(append([]string{"id"}, groupLessonsFieldNames...)...)

	for _, dbA := range dbGroupLessons {
		query = query.Values(dbA.values()...)
	}

	query = query.Suffix("ON CONFLICT ON CONSTRAINT group_lesson_unique DO NOTHING").PlaceholderFormat(squirrel.Dollar)

	sql, bound, err := query.ToSql()
	if err != nil {
		return err
	}

	if _, err = d.db.ExecContext(ctx, sql, bound...); err != nil {
		return fmt.Errorf("cannot insert query: %v, args %v, into %v: %w", query, bound, groupLessonTable, err)
	}

	return nil
}
