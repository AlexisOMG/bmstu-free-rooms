package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Lesson struct {
	ID          string
	Name        string
	TeacherName *string
	Kind        *string
}

type LessonFilters struct {
	Name *string
}

func (s *Service) SaveLessons(ctx context.Context, lessons ...Lesson) ([]string, error) {
	lessonsToSave := make([]Lesson, 0, len(lessons))
	lessonsIDs := make([]string, 0, len(lessons))

	for _, l := range lessons {
		l.ID = uuid.NewString()
		lessonsIDs = append(lessonsIDs, l.ID)
		lessonsToSave = append(lessonsToSave, l)
	}

	if err := s.scheduleStorage.SaveLessons(ctx, lessonsToSave...); err != nil {
		return []string{}, fmt.Errorf("cannot save lessons: %w", err)
	}

	return lessonsIDs, nil
}

func (s *Service) ListLessons(ctx context.Context, filters *LessonFilters) ([]Lesson, error) {
	return s.scheduleStorage.ListLessons(ctx, filters)
}
