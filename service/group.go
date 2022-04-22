package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Group struct {
	ID   string
	Name string
}

type GroupFilters struct {
	Name *string
}

func (s *Service) SaveGroups(ctx context.Context, groups ...Group) ([]string, error) {
	groupsToSave := make([]Group, 0, len(groups))
	groupsIDs := make([]string, 0, len(groups))

	for _, g := range groups {
		g.ID = uuid.NewString()
		groupsIDs = append(groupsIDs, g.ID)
		groupsToSave = append(groupsToSave, g)
	}

	if err := s.scheduleStorage.SaveGroups(ctx, groupsToSave...); err != nil {
		return []string{}, fmt.Errorf("cannot save groups: %w", err)
	}

	return groupsIDs, nil
}

func (s *Service) ListGroups(ctx context.Context, filters *GroupFilters) ([]Group, error) {
	return s.scheduleStorage.ListGroups(ctx, filters)
}

type GroupLesson struct {
	ID       string
	GroupID  string
	LessonID string
}

func (s *Service) SaveGroupLessons(ctx context.Context, gls ...GroupLesson) ([]string, error) {
	glsToSave := make([]GroupLesson, 0, len(gls))
	glIDs := make([]string, 0, len(gls))

	for _, gl := range gls {
		gl.ID = uuid.NewString()
		glIDs = append(glIDs, gl.ID)
		glsToSave = append(glsToSave, gl)
	}

	if err := s.scheduleStorage.SaveGroupLessons(ctx, glsToSave...); err != nil {
		return []string{}, fmt.Errorf("cannot save group_lessons: %w", err)
	}

	return glIDs, nil
}
