package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

var (
	suffixMapping = map[string]string{
		"л":  "УЛК",
		"а":  "УЛК",
		"б":  "УЛК",
		"ю":  "ГЗ",
		"аю": "ГЗ",
	}
)

type Audience struct {
	ID       string
	Number   string
	Building string
	Floor    int
	Suffix   *string
}

func (a *Audience) fillCalculatedFields() error {
	if a.Suffix == nil {
		a.Building = "ГЗ"
	} else {
		building, ok := suffixMapping[*a.Suffix]
		if !ok {
			return fmt.Errorf("unknown building suffix: %s", *a.Suffix)
		}
		a.Building = building
	}

	if len(a.Number) > 0 {
		floor, err := strconv.Atoi(a.Number[:1])
		if err != nil {
			return fmt.Errorf("invalid audience number %s: %w", a.Number, err)
		}
		a.Floor = floor
	} else {
		return &ValidationError{
			ObjectKind: "Audience",
			Message:    "empty audience number",
		}
	}

	return nil
}

func (s *Service) SaveAudiences(ctx context.Context, audiences ...Audience) ([]string, error) {
	audiencesToSave := make([]Audience, 0, len(audiences))
	audiencesIDs := make([]string, 0, len(audiences))

	for _, a := range audiences {
		a.ID = uuid.NewString()
		if err := a.fillCalculatedFields(); err != nil {
			return []string{}, err
		}

		audiencesIDs = append(audiencesIDs, a.ID)
		audiencesToSave = append(audiencesToSave, a)
	}

	if err := s.scheduleStorage.SaveAudiences(ctx, audiencesToSave...); err != nil {
		return []string{}, fmt.Errorf("cannot save audiences: %w", err)
	}

	return audiencesIDs, nil
}

func (s *Service) ListAudienceByNumber(ctx context.Context, number string, suffix *string) (Audience, error) {
	return s.scheduleStorage.ListAudienceByNumber(ctx, number, suffix)
}

type EmptyAudiencesFilter struct {
	Building string
	WeekType string
	WeekDay  string
	Period   int
	Floor    int
}

func (s *Service) ListEmptyAudiences(ctx context.Context, filters *EmptyAudiencesFilter) ([]Audience, error) {
	return s.scheduleStorage.ListEmptyAudiences(ctx, filters)
}
