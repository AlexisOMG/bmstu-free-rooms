package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/AlexisOMG/bmstu-free-rooms/database"
	"github.com/AlexisOMG/bmstu-free-rooms/handlers"
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
	needDownload := flag.Bool("p", false, "use it for downloading schedule")
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

	if needDownload != nil && *needDownload {
		downloader := handlers.NewICSDownloader(*conf.ScheduleDir)
		err = downloader.DownloadICS(ctx)
		if err != nil {
			logger.WithError(err).Fatal("ics loading failed")
		}
		logger.Info("loaded ics files")
		err = icsparser.ProcessICSFiles(ctx, srvc, *conf.ScheduleDir)
		if err != nil {
			logger.WithError(err).Fatal("ics processing failed")
		}
		logger.Info("processed ics files")
	}

	bot := handlers.NewBot(*conf.Token)

	bot.Listen(ctx, srvc)
}
