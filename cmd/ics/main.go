package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	srvc := service.NewService(storage)

	type Query struct {
		MessageID int
		Filter    service.EmptyAudiencesFilter
	}

	queries := make(map[int64]Query)

	bot, err := tgbotapi.NewBotAPI(*conf.Token)
	if err != nil {
		logger.WithError(err).Fatal("cannot connect to bot")
	}

	// bot.Debug = true

	u := tgbotapi.NewUpdate(583350088)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// go func(ctx context.Context, w *sync.WaitGroup, updates tgbotapi.UpdatesChannel) {
	// 	defer w.Done()
	// 	for upd := range updates {
	// 		fmt.Println(upd.UpdateID, upd.Message.Text)
	// 		if _, ok := <-ctx.Done(); !ok {
	// 			return
	// 		}
	// 	}
	// }(ctx, wg, updates)

	// wg.Wait()
	// В канал updates будут приходить все новые сообщения.
	for update := range updates {
		fmt.Println("UPDATE ID: ", update.UpdateID)
		if update.Message != nil {
			if update.Message.Text != "" && update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "День недели")
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
					{tgbotapi.NewInlineKeyboardButtonData("Sunday", "Sunday")},
					{tgbotapi.NewInlineKeyboardButtonData("Monday", "Monday")},
					{tgbotapi.NewInlineKeyboardButtonData("Tuesday", "Tuesday")},
					{tgbotapi.NewInlineKeyboardButtonData("Wednesday", "Wednesday")},
					{tgbotapi.NewInlineKeyboardButtonData("Thursday", "Thursday")},
					{tgbotapi.NewInlineKeyboardButtonData("Friday", "Friday")},
					{tgbotapi.NewInlineKeyboardButtonData("Saturday", "Saturday")},
				}
				msg.ReplyMarkup = keyboard
				cm, err := bot.Send(msg)
				if err != nil {
					logger.WithError(err).Fatal("cannot send msg to bot")
				}
				queries[update.Message.Chat.ID] = Query{
					MessageID: cm.MessageID,
				}
			} else {
				logger.WithField("unknown msg", update.Message).Warning()
			}
		} else if update.CallbackQuery != nil {
			clq := update.CallbackQuery
			switch clq.Message.Text {
			case "День недели":
				query := queries[clq.Message.Chat.ID]
				query.Filter.WeekDay = clq.Data
				queries[clq.Message.Chat.ID] = query

				// msg := tgbotapi.NewMessage(clq.Message.Chat.ID, "Числитель или Знаменатель")
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
					{tgbotapi.NewInlineKeyboardButtonData("ЧС", "ЧС")},
					{tgbotapi.NewInlineKeyboardButtonData("ЗН", "ЗН")},
				}
				// msg.ReplyMarkup = keyboard
				msg := tgbotapi.NewEditMessageTextAndMarkup(clq.Message.Chat.ID, query.MessageID, "Числитель или Знаменатель", keyboard)
				if _, err := bot.Send(msg); err != nil {
					logger.WithError(err).Fatal("cannot send msg to bot")
				}
				fmt.Println(queries)
			case "Числитель или Знаменатель":
				query := queries[clq.Message.Chat.ID]
				query.Filter.WeekType = clq.Data
				queries[clq.Message.Chat.ID] = query

				// msg := tgbotapi.NewMessage(clq.Message.Chat.ID, "Корпус")
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
					{tgbotapi.NewInlineKeyboardButtonData("ГЗ", "ГЗ")},
					{tgbotapi.NewInlineKeyboardButtonData("УЛК", "УЛК")},
				}
				// msg.ReplyMarkup = keyboard
				msg := tgbotapi.NewEditMessageTextAndMarkup(clq.Message.Chat.ID, query.MessageID, "Корпус", keyboard)
				if _, err := bot.Send(msg); err != nil {
					logger.WithError(err).Fatal("cannot send msg to bot")
				}
				fmt.Println(queries)
			case "Корпус":
				query := queries[clq.Message.Chat.ID]
				query.Filter.Building = clq.Data
				queries[clq.Message.Chat.ID] = query

				// msg := tgbotapi.NewMessage(clq.Message.Chat.ID, "Этаж")
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				if clq.Data == "ГЗ" {
					for i := 1; i <= 5; i++ {
						keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
							tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i), strconv.Itoa(i)),
						})
					}
				} else if clq.Data == "УЛК" {
					for i := 1; i <= 11; i++ {
						keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
							tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i), strconv.Itoa(i)),
						})
					}
				}
				// msg.ReplyMarkup = keyboard
				msg := tgbotapi.NewEditMessageTextAndMarkup(clq.Message.Chat.ID, query.MessageID, "Этаж", keyboard)
				if _, err := bot.Send(msg); err != nil {
					logger.WithError(err).Fatal("cannot send msg to bot")
				}
				fmt.Println(queries)
			case "Этаж":
				query := queries[clq.Message.Chat.ID]
				floor, err := strconv.Atoi(clq.Data)
				if err != nil {
					logger.WithError(err).Fatal("cannot convert floor")
				}
				query.Filter.Floor = floor
				queries[clq.Message.Chat.ID] = query

				// msg := tgbotapi.NewMessage(clq.Message.Chat.ID, "Пара")
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				for i := 1; i <= 7; i++ {
					keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
						tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i), strconv.Itoa(i)),
					})
				}
				// msg.ReplyMarkup = keyboard
				msg := tgbotapi.NewEditMessageTextAndMarkup(clq.Message.Chat.ID, query.MessageID, "Пара", keyboard)
				if _, err := bot.Send(msg); err != nil {
					logger.WithError(err).Fatal("cannot send msg to bot")
				}
				fmt.Println(queries)
			case "Пара":
				query := queries[clq.Message.Chat.ID]
				period, err := strconv.Atoi(clq.Data)
				if err != nil {
					logger.WithError(err).Fatal("cannot convert period")
				}
				query.Filter.Period = period
				auds, err := srvc.ListEmptyAudiences(ctx, &query.Filter)
				if err != nil {
					logger.WithError(err).Fatal("cannot list empty audiences")
				}
				if len(auds) == 0 {
					msg := tgbotapi.NewMessage(clq.Message.Chat.ID, "No Empty Audiences")
					if _, err := bot.Send(msg); err != nil {
						logger.WithError(err).Fatal("cannot send msg to bot")
					}
					delete(queries, clq.Message.Chat.ID)
				} else {
					resp := ""
					for _, aud := range auds {
						resp += aud.Number
						if aud.Suffix != nil {
							resp += *aud.Suffix
						}
						resp += " "
					}
					msg := tgbotapi.NewMessage(clq.Message.Chat.ID, resp)
					if _, err := bot.Send(msg); err != nil {
						logger.WithError(err).Fatal("cannot send msg to bot")
					}
					delete(queries, clq.Message.Chat.ID)
				}
				msg := tgbotapi.NewMessage(clq.Message.Chat.ID, "День недели")
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				keyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
					{tgbotapi.NewInlineKeyboardButtonData("Sunday", "Sunday")},
					{tgbotapi.NewInlineKeyboardButtonData("Monday", "Monday")},
					{tgbotapi.NewInlineKeyboardButtonData("Tuesday", "Tuesday")},
					{tgbotapi.NewInlineKeyboardButtonData("Wednesday", "Wednesday")},
					{tgbotapi.NewInlineKeyboardButtonData("Thursday", "Thursday")},
					{tgbotapi.NewInlineKeyboardButtonData("Friday", "Friday")},
					{tgbotapi.NewInlineKeyboardButtonData("Saturday", "Saturday")},
				}
				msg.ReplyMarkup = keyboard
				cm, err := bot.Send(msg)
				if err != nil {
					logger.WithError(err).Fatal("cannot send msg to bot")
				}
				queries[clq.Message.Chat.ID] = Query{
					MessageID: cm.MessageID,
				}
			default:
				logger.WithField("unknown query", clq).Fatal()
			}
		} else {
			logger.WithField("unknown upd", update).Fatal()
		}
		// Создав структуру - можно её отправить обратно боту
		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID
		// bot.Send(msg)
	}

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
