package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/qwsnxnjene/telegram-bot/lib/e"
	"github.com/qwsnxnjene/telegram-bot/storage"
)

const (
	RndCmd      = "/rnd"
	HelpCmd     = "/help"
	StartCmd    = "/start"
	LastFiveCmd = "/get"
)

func (p *Processor) doCmd(text string, chatID int, userName string) error {
	text = strings.TrimSpace(text)

	log.Printf("[doCmd]: got new command '%s' from '%s'", text, userName)

	if isAddCmd(text) {
		return p.savePage(chatID, text, userName)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, userName)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	case LastFiveCmd:
		return p.sendLastFive(chatID, userName)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor) sendLastFive(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("[sendTen]: can't do command: can't send last 10 pages", err) }()

	pages, err := p.storage.PickLastFive(context.Background(), username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	var message string
	for i, page := range pages {
		message += fmt.Sprintf("%d. %s\n\n", i+1, page.URL)
	}
	if len(message) == 0 {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	message = message[:len(message)-1]

	if err := p.tg.SendMessage(chatID, message); err != nil {
		return err
	}

	for _, page := range pages {
		if err := p.storage.Remove(context.Background(), page); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("[savePage]: can't do command: save page", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(context.Background(), page)
	if err != nil {
		return err
	}
	if isExists {
		return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("[sendRandom]: can't do command: can't send random page", err) }()

	page, err := p.storage.PickRandom(context.Background(), username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(context.Background(), page)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
