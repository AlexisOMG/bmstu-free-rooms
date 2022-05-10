package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/AlexisOMG/bmstu-free-rooms/database"
	"github.com/AlexisOMG/bmstu-free-rooms/icsparser"
	"github.com/AlexisOMG/bmstu-free-rooms/service"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancel()

	logger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	ctx = context.WithValue(ctx, "logger", logger)

	configPath := flag.String("c", "config.yaml", "path to your config")
	flag.Parse()

	conf, err := readConfig(*configPath)
	if err != nil {
		logger.WithError(err).Fatal("failed to read config")
	}

	storage, err := database.NewDatabase(ctx, conf.Database)
	if err != nil {
		logger.WithError(err).Fatal("failed to create database")
	}
	defer storage.Close(ctx)

	err = storage.Ping(ctx)
	if err != nil {
		logger.WithError(err).Fatal("database ping failed")
	}
	logger.Info("connected to database")

	srvc := service.NewService(storage)

	// d, err := icsparser.ParseICS(ctx, "schedules/ИУ9-62Б.ics")

	// err = icsparser.SaveData(ctx, srvc, d)
	// if err != nil {
	// 	logger.WithError(err).Fatal("ics parse failed")
	// }

	files, err := ioutil.ReadDir(*conf.ScheduleDir)
	if err != nil {
		logger.WithError(err).Fatal("schedules dir reading failed")
	}

	schedules := make([]string, 0, 1024)

	for _, f := range files {
		if !f.IsDir() {
			schedules = append(schedules, *conf.ScheduleDir+"/"+f.Name())
		}
	}

	for _, s := range schedules {
		fmt.Println("PROCESSING: ", s)
		d, err := icsparser.ParseICS(ctx, s)
		if err != nil {
			logger.WithError(err).Fatal("ics parse failed")
		}
		err = icsparser.SaveData(ctx, srvc, d)
		if err != nil {
			logger.WithError(err).Fatal("ics save failed")
		}
		fmt.Println("PROCESSED: ", s)
	}

	// fmt.Println(schedules, len(schedules))

	// icsFiles := make(chan string, len(schedules))
	// errs := make(chan error)

	// for _, s := range schedules {
	// 	icsFiles <- s
	// }

	// wg := sync.WaitGroup{}

	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(w *sync.WaitGroup, in chan string, errs chan error) {
	// 		defer w.Done()
	// 		for {
	// 			m, ok := <-in
	// 			if !ok {
	// 				return
	// 			}
	// 			d, err := icsparser.ParseICS(ctx, m)
	// 			if err != nil {
	// 				errs <- err
	// 				return
	// 			}
	// 			if err := icsparser.SaveData(ctx, srvc, d); err != nil {
	// 				errs <- err
	// 				return
	// 			}
	// 		}
	// 	}(&wg, icsFiles, errs)
	// }

	// go func() {
	// 	wg.Wait()
	// 	close(errs)
	// }()

	// for err := range errs {
	// 	logger.WithError(err).Fatal("failed to process ics")
	// }

	fmt.Println("DONE")
}
