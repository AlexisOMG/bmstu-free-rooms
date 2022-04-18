package service

import "context"

type ScheduleStorage interface {
	SaveUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context, filters *UserFilters) ([]User, error)

	SaveAudiences(ctx context.Context, audiences ...Audience) error

	SaveLessons(ctx context.Context, lessons ...Lesson) error
	ListLessons(ctx context.Context, filters *LessonFilters) ([]Lesson, error)
}
