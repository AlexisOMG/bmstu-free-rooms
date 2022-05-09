package icsparser

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
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
		res.Schedules = append(res.Schedules, s)
		// fmt.Println()
	}
	return res, nil
}
