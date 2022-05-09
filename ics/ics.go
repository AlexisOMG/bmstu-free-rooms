package ics

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

func ParseICS(ctx context.Context, path string) error {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err.Error())
	}

	cal, err := ics.ParseCalendar(strings.NewReader(string(d)))

	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == "X-WR-CALNAME" {
			fmt.Println(prop.Value)
		}
	}

	for _, comp := range cal.Components {
		fmt.Println(comp.UnknownPropertiesIANAProperties())
		for _, prop := range comp.UnknownPropertiesIANAProperties() {
			switch prop.IANAToken {
			case "SUMMARY":
				fmt.Print(prop.Value)
			case "DTSTART":
				ts := prop.Value
				start, err := time.Parse("20060102T150405Z", ts)
				if err != nil {
					log.Fatal(err.Error())
				}
				fmt.Printf(" Weekday: %s", start.Weekday().String())
			case "RRULE":
				rules := strings.Split(prop.Value, ";")
				for _, r := range rules {
					if strings.Contains(r, "INTERVAL") {
						fmt.Printf(" INTERVAL: %s", strings.Split(r, "=")[1])
					}
				}
			}
		}
		fmt.Println()
	}
	return nil
}
