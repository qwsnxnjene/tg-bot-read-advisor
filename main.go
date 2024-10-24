package main

import (
	"context"
	"errors"
	"flag"
	"log"

	"github.com/qwsnxnjene/telegram-bot/clients/telegram"
	eventconsumer "github.com/qwsnxnjene/telegram-bot/consumer/event-consumer"
	tg "github.com/qwsnxnjene/telegram-bot/events/telegram"
	"github.com/qwsnxnjene/telegram-bot/storage/sqlite"
)

const (
	tgBotHost         = "api.telegram.org"
	storageSqlitePath = "data/sqlite/database.db"
	batchSize         = 1
)

/*
TODO: 1. Добавить хранение заголовков статей в БД
2. Добавить заголовки в ответы
3. Добавить дату добавления статьи
*/

func main() {

	s, err := sqlite.New(storageSqlitePath)
	if err != nil {
		log.Fatal("[main]: can't connect to storage: ", err)
	}

	err = s.Init(context.TODO())
	if err != nil {
		log.Fatal("[main]: can't init storage: ", err)
	}

	tgClient := telegram.New(tgBotHost, mustToken())
	eventsProcessor := tg.New(tgClient, s)

	log.Print("service started")

	consumer := eventconsumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("[main]: service is stopped", err)
	}
}

func mustToken() string {
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal(errors.New("empty token"))
	}

	return *token
}
