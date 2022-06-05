package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

type ICSDownloader interface {
	DownloadICS(ctx context.Context) error
}

func NewICSDownloader(scheduleDir string) ICSDownloader {
	return &downloader{
		pathToDir: scheduleDir,
	}
}

type downloader struct {
	pathToDir string
}

func (d *downloader) DownloadICS(ctx context.Context) error {
	logger := ctx.Value("logger").(*logrus.Logger)

	refs, err := getAllScheduleRefs()
	if err != nil {
		return err
	}

	for ref, group := range refs {
		resp, err := http.Get(ref)
		if err != nil {
			logger.WithError(err).Warning("cannot download schedule")
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			logger.WithError(fmt.Errorf("bad status code: %d, %s, %s", resp.StatusCode, ref, group)).Warning("cannot download schedule")
			continue
		}

		out, err := os.Create(d.pathToDir + "/" + group + ".ics")
		if err != nil {
			logger.WithError(err).Warning("cannot create ics file")
			continue
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			logger.WithError(err).Warning("cannot copy ics file")
		}
	}
	return nil
}

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

func getAllScheduleRefs() (ScheduleRef, error) {
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

	elements := getElementsByClass(node, className)
	for _, elem := range elements {

		for child := elem.FirstChild; child != nil; child = child.NextSibling {
			file := "https://lks.bmstu.ru"
			skip := false
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
			}
		}
	}

	return res, nil
}
