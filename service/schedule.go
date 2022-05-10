package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	ID         string
	AudienceID string
	LessonID   string
	WeekType   string
	WeekDay    string
	Start      *time.Time
	End        *time.Time
	Period     int
}

func (s *Schedule) computePeriod() error {
	if s.Start == nil {
		return &ValidationError{
			ObjectKind: "Schedule",
			Message:    "nil start time",
		}
	}
	if s.End == nil {
		return &ValidationError{
			ObjectKind: "Schedule",
			Message:    "nil end time",
		}
	}

	hStart, mStart, _ := s.Start.Clock()
	hEnd, mEnd, _ := s.End.Clock()

	switch {
	case hStart == 5 && mStart == 30 && hEnd == 7 && mEnd == 5:
		s.Period = 1
	case hStart == 7 && mStart == 15 && hEnd == 8 && mEnd == 50:
		s.Period = 2
	case hStart == 9 && mStart == 0 && hEnd == 10 && mEnd == 35:
		s.Period = 3
	case hStart == 10 && mStart == 50 && hEnd == 12 && mEnd == 25:
		s.Period = 4
	case hStart == 12 && mStart == 40 && hEnd == 14 && mEnd == 15:
		s.Period = 5
	case hStart == 14 && mStart == 25 && hEnd == 16 && mEnd == 0:
		s.Period = 6
	case hStart == 16 && mStart == 10 && hEnd == 17 && mEnd == 45:
		s.Period = 7
	default:
		return &ValidationError{
			ObjectKind: "Schedule",
			Message:    "invalid start end times",
		}
	}

	locStart := s.Start.Add(time.Hour * 3)
	s.Start = &locStart

	locEnd := s.End.Add(time.Hour * 3)
	s.End = &locEnd

	return nil
}

type ScheduleFilters struct {
}

func (s *Service) SaveSchedules(ctx context.Context, schedules ...Schedule) ([]string, error) {
	schedulesToSave := make([]Schedule, 0, len(schedules))
	schedulesIDs := make([]string, 0, len(schedules))

	for _, s := range schedules {
		err := s.computePeriod()
		if err != nil {
			return []string{}, err
		}
		s.ID = uuid.NewString()
		schedulesIDs = append(schedulesIDs, s.ID)
		schedulesToSave = append(schedulesToSave, s)
	}

	if err := s.scheduleStorage.SaveSchedules(ctx, schedulesToSave...); err != nil {
		return []string{}, fmt.Errorf("cannot save schedules: %w", err)
	}

	return schedulesIDs, nil
}

func (s *Service) ListSchedules(ctx context.Context, filters *ScheduleFilters) ([]Schedule, error) {
	return s.scheduleStorage.ListSchedules(ctx, filters)
}
