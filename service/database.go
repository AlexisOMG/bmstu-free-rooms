package service

import "context"

type ScheduleStorage interface {
	SaveUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context, filters *UserFilters) ([]User, error)

	SaveAudiences(ctx context.Context, audiences ...Audience) error

	SaveLessons(ctx context.Context, lessons ...Lesson) error
	ListLessons(ctx context.Context, filters *LessonFilters) ([]Lesson, error)

	SaveGroups(ctx context.Context, groups ...Group) error
	ListGroups(ctx context.Context, filters *GroupFilters) ([]Group, error)

	SaveGroupLessons(ctx context.Context, gls ...GroupLesson) error

	SaveSchedules(ctx context.Context, lessons ...Schedule) error
	ListSchedules(ctx context.Context, filters *ScheduleFilters) ([]Schedule, error)

	ListEmptyAudiences(ctx context.Context, filters *EmptyAudiencesFilter) ([]Audience, error)
}
