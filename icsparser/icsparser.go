package icsparser

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/sirupsen/logrus"

	"github.com/AlexisOMG/bmstu-free-rooms/service"
)

type Schedule struct {
	Name     string
	Start    *time.Time
	End      *time.Time
	Interval string
	Location string
	Teacher  string
}

type Data struct {
	Group     string
	Schedules []Schedule
}

var (
	milReg      = regexp.MustCompile(`(?i)вуц`)
	cafReg      = regexp.MustCompile(`(?i)каф`)
	peReg       = regexp.MustCompile(`(?i)Элективный курс по физической культуре и спорту`)
	ulkReg      = regexp.MustCompile(`^[\d\.]+[лаб]$`)
	gzReg       = regexp.MustCompile(`^[\d\.]+(ю|аю)?$`)
	suffixReg   = regexp.MustCompile(`[а-яА-Я]+`)
	scheduleReg = regexp.MustCompile(`^Расписание `)
)

func ParseICS(ctx context.Context, path string) (Data, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return Data{}, err
	}

	res := Data{}

	cal, err := ics.ParseCalendar(strings.NewReader(string(d)))

	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == "X-WR-CALNAME" {
			// fmt.Println(prop.Value)
			res.Group = prop.Value
		}
	}

	for _, comp := range cal.Components {
		s := Schedule{}
		// fmt.Println(comp.UnknownPropertiesIANAProperties())
		for _, prop := range comp.UnknownPropertiesIANAProperties() {
			switch prop.IANAToken {
			case "SUMMARY":
				// fmt.Print(" ", prop.Value, " ")
				s.Name = prop.Value
			case "DTSTART":
				ts := prop.Value
				start, err := time.Parse("20060102T150405Z", ts)
				if err != nil {
					return Data{}, err
				}
				s.Start = &start
				// fmt.Printf(" Start Weekday: %s ", start.String())
			case "DTEND":
				ts := prop.Value
				end, err := time.Parse("20060102T150405Z", ts)
				if err != nil {
					return Data{}, err
				}
				s.End = &end
				// fmt.Printf(" End Weekday: %s ", start.String())
			case "RRULE":
				rules := strings.Split(prop.Value, ";")
				for _, r := range rules {
					if strings.Contains(r, "INTERVAL") {
						s.Interval = strings.Split(r, "=")[1]
						// fmt.Printf(" INTERVAL: %s ", strings.Split(r, "=")[1])
					}
				}
			case "LOCATION":
				// fmt.Print(" ", prop.Value, " ")
				s.Location = prop.Value
			case "DESCRIPTION":
				// fmt.Print(" ", prop.Value, " ")
				s.Teacher = prop.Value
			}
		}

		if s.Name != "" && s.Location != "" && s.Start != nil && s.End != nil &&
			!milReg.MatchString(s.Name) && !cafReg.MatchString(s.Location) && !peReg.MatchString(s.Name) &&
			(ulkReg.MatchString(s.Location) || gzReg.MatchString(s.Location)) {
			res.Schedules = append(res.Schedules, s)
		}
		// fmt.Println()
	}
	return res, nil
}

func SaveData(ctx context.Context, srvc *service.Service, data Data) error {
	log := ctx.Value("logger").(*logrus.Logger)
	var groupID string
	if loc := scheduleReg.FindStringIndex(data.Group); loc != nil {
		groupName := data.Group[loc[1]:]

		gs, err := srvc.ListGroups(ctx, &service.GroupFilters{Name: &groupName})
		if err != nil {
			if errors.Is(err, service.ErrorNotFound) {
				ids, err := srvc.SaveGroups(ctx, service.Group{Name: groupName})
				if err != nil {
					return err
				}
				gs = []service.Group{
					{ID: ids[0]},
				}
			} else {
				return err
			}
		}
		groupID = gs[0].ID
	} else {
		return fmt.Errorf("invalid group name: %s", data.Group)
	}

	audienceIDs := make(map[string]string)
	lessonIDs := make(map[string]string)
	schs := map[string]map[int][]Schedule{
		"Sunday":    make(map[int][]Schedule),
		"Monday":    make(map[int][]Schedule),
		"Tuesday":   make(map[int][]Schedule),
		"Wednesday": make(map[int][]Schedule),
		"Thursday":  make(map[int][]Schedule),
		"Friday":    make(map[int][]Schedule),
		"Saturday":  make(map[int][]Schedule),
	}

	for _, schedule := range data.Schedules {
		if _, ok := audienceIDs[schedule.Location]; !ok {
			var number string
			var suffix *string
			if loc := suffixReg.FindStringIndex(schedule.Location); loc != nil {
				number = schedule.Location[:loc[0]]
				suf := schedule.Location[loc[0]:loc[1]]
				suffix = &suf
			} else {
				number = schedule.Location
			}

			aud, err := srvc.ListAudienceByNumber(ctx, number, suffix)
			if err != nil {
				if errors.Is(err, service.ErrorNotFound) {
					ids, err := srvc.SaveAudiences(ctx, service.Audience{
						Number: number,
						Suffix: suffix,
					})
					if err != nil {
						return err
					}
					aud.ID = ids[0]
				} else {
					return err
				}
			}
			audienceIDs[schedule.Location] = aud.ID
		}
		if _, ok := lessonIDs[schedule.Name]; !ok {
			lesson := service.Lesson{Name: schedule.Name}
			if schedule.Teacher != "" {
				lesson.TeacherName = &schedule.Teacher
			}
			ids, err := srvc.SaveLessons(ctx, lesson)
			if err != nil {
				return err
			}
			lessonIDs[schedule.Name] = ids[0]
		}

		hStart, mStart, _ := schedule.Start.Clock()
		hEnd, mEnd, _ := schedule.End.Clock()
		period := -1

		switch {
		case hStart == 5 && mStart == 30 && hEnd == 7 && mEnd == 5:
			period = 1
		case hStart == 7 && mStart == 15 && hEnd == 8 && mEnd == 50:
			period = 2
		case hStart == 9 && mStart == 0 && hEnd == 10 && mEnd == 35:
			period = 3
		case hStart == 10 && mStart == 50 && hEnd == 12 && mEnd == 25:
			period = 4
		case hStart == 12 && mStart == 40 && hEnd == 14 && mEnd == 15:
			period = 5
		case hStart == 14 && mStart == 25 && hEnd == 16 && mEnd == 0:
			period = 6
		case hStart == 16 && mStart == 10 && hEnd == 17 && mEnd == 45:
			period = 7
		default:
			log.WithError(fmt.Errorf("invalid start end: %v", schedule)).Warning("skip schedule")
			continue
		}

		schs[schedule.Start.Weekday().String()][period] = append(schs[schedule.Start.Weekday().String()][period], schedule)
	}

	for _, schedule := range data.Schedules {
		lessonID, ok := lessonIDs[schedule.Name]
		if !ok {
			return fmt.Errorf("unknown lesson: %s, group: %s", schedule.Name, data.Group)
		}
		_, err := srvc.SaveGroupLessons(ctx, service.GroupLesson{
			GroupID:  groupID,
			LessonID: lessonID,
		})
		if err != nil {
			return err
		}
	}

	for weekday := range schs {
		for _, schedules := range schs[weekday] {
			switch len(schedules) {
			case 0:
				continue
			case 1:
				s := service.Schedule{
					WeekType: "ЧС",
					WeekDay:  weekday,
					Start:    schedules[0].Start,
					End:      schedules[0].End,
				}

				if id, ok := audienceIDs[schedules[0].Location]; ok {
					s.AudienceID = id
				} else {
					return fmt.Errorf("unknown audince: %v", schedules[0])
				}

				if id, ok := lessonIDs[schedules[0].Name]; ok {
					s.LessonID = id
				} else {
					return fmt.Errorf("unknown lesson: %v", schedules[0])
				}

				_, err := srvc.SaveSchedules(ctx, s)
				if err != nil {
					return err
				}

				s.WeekType = "ЗН"

				_, err = srvc.SaveSchedules(ctx, s)
				if err != nil {
					return err
				}
			case 2:
				s1 := service.Schedule{
					WeekDay: weekday,
					Start:   schedules[0].Start,
					End:     schedules[0].End,
				}

				if id, ok := audienceIDs[schedules[0].Location]; ok {
					s1.AudienceID = id
				} else {
					return fmt.Errorf("unknown audince: %v", schedules[0])
				}

				if id, ok := lessonIDs[schedules[0].Name]; ok {
					s1.LessonID = id
				} else {
					return fmt.Errorf("unknown lesson: %v", schedules[0])
				}

				s2 := service.Schedule{
					WeekDay: weekday,
					Start:   schedules[1].Start,
					End:     schedules[1].End,
				}

				if id, ok := audienceIDs[schedules[1].Location]; ok {
					s2.AudienceID = id
				} else {
					return fmt.Errorf("unknown audince: %v", schedules[1])
				}

				if id, ok := lessonIDs[schedules[1].Name]; ok {
					s2.LessonID = id
				} else {
					return fmt.Errorf("unknown lesson: %v", schedules[1])
				}

				if s1.Start.Before(*s2.Start) {
					s1.WeekType = "ЧС"
					s2.WeekType = "ЗН"
				} else {
					s2.WeekType = "ЧС"
					s1.WeekType = "ЗН"
				}
				_, err := srvc.SaveSchedules(ctx, s1, s2)
				if err != nil {
					return fmt.Errorf("cannot save 2 schedules: %w", err)
				}
			default:
				return fmt.Errorf("smth went wrong: %v", data)
			}
		}
	}

	return nil
}
