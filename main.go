package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"golang.org/x/net/html"
)

func isNeeded(node *html.Node, class string) bool {
	var names []string
	for _, attr := range node.Attr {
		if attr.Key == "class" {
			names = strings.Split(attr.Val, " ")
			break
		}
	}
	for _, name := range names {
		if name == class {
			return true
		}
	}
	return false
}

func getElementsByClass(node *html.Node, class string) []*html.Node {
	var elements []*html.Node
	if isNeeded(node, class) {
		elements = append(elements, node)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		elements = append(elements, getElementsByClass(child, class)...)
	}
	return elements
}

type ScheduleRef map[string]string

func GetAllScheduleRefs() (ScheduleRef, error) {
	className := "col-xs-10"
	url := "https://lks.bmstu.ru/schedule/list"
	res := make(ScheduleRef)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Cannot get html: %w", err)
	}
	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse html: %w", err)
	}
	// var refs []ScheduleRef
	elements := getElementsByClass(node, className)
	for _, elem := range elements {

		for child := elem.FirstChild; child != nil; child = child.NextSibling {
			file := "https://lks.bmstu.ru"
			skip := false
			// fmt.Println(child.Attr)
			for _, attr := range child.Attr {
				switch attr.Key {
				case "href":
					file += attr.Val
				case "title":
					if attr.Val == "нет расписания" {
						skip = true
					}
				}
			}
			if !skip && file != "https://lks.bmstu.ru" {
				res[file+".ics"] = strings.TrimSpace(child.FirstChild.Data)
				// refs = append(refs, ScheduleRef{
				// 	Ref:   file + ".ics",
				// 	Title: strings.TrimSpace(child.FirstChild.Data),
				// })
			}
		}
	}

	return res, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("wrong usage")
	}

	d, err := ioutil.ReadFile(os.Args[1])
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

	// refs, err := GetAllScheduleRefs()
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// fmt.Println(len(refs))

	// wg := sync.WaitGroup{}

	// errs := make(chan error)
	// cnt := 0

	// for ref, group := range refs {
	// 	cnt += 1
	// 	wg.Add(1)
	// 	go func(r, g string, waiter *sync.WaitGroup, errs chan error) {
	// 		defer waiter.Done()
	// 		resp, err := http.Get(r)
	// 		if err != nil {
	// 			errs <- err
	// 			return
	// 		}
	// 		defer resp.Body.Close()
	// 		if resp.StatusCode != 200 {
	// 			errs <- fmt.Errorf("unknown error: %d, %s, %s", resp.StatusCode, r, g)
	// 			return
	// 		}

	// 		out, err := os.Create("schedules/" + g + ".ics")
	// 		if err != nil {
	// 			errs <- err
	// 			return
	// 		}
	// 		defer out.Close()

	// 		_, err = io.Copy(out, resp.Body)
	// 		errs <- err
	// 		return
	// 	}(ref, group, &wg, errs)
	// }
	// go func() {
	// 	wg.Wait()
	// 	close(errs)
	// }()

	// for err := range errs {
	// 	fmt.Println(err)
	// }
	// fmt.Println("DONE")
}
