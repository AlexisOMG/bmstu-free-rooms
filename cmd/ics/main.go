package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/AlexisOMG/bmstu-free-rooms/database"
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

	_ = service.NewService(storage)

	// audiences, err := srvc.ListEmptyAudiences(ctx, &service.EmptyAudiencesFilter{
	// 	Building: "ГЗ",
	// 	WeekType: "ЧС",
	// 	WeekDay:  "Tuesday",
	// 	Period:   1,
	// 	Floor:    3,
	// })
	// if err != nil {
	// 	logger.WithError(err).Fatal("ListEmptyAudiences failed")
	// }
	// fmt.Println(audiences)
}
