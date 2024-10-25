package storage

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/qwsnxnjene/telegram-bot/lib/e"
)

type Storage interface {
	Save(ctx context.Context, p *Page) error
	PickRandom(ctx context.Context, userName string) (*Page, error)
	PickLastFive(ctx context.Context, userName string) ([]*Page, error)
	Remove(ctx context.Context, p *Page) error
	IsExists(ctx context.Context, p *Page) (bool, error)
}

type Page struct {
	URL      string
	UserName string
	Title    string
	Date     string
}

func (p Page) Hash() (string, error) {
	h := sha1.New()
	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", e.Wrap("[Hash]: can't calculate hash", err)
	}

	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", e.Wrap("[Hash]: can't calculate hash", err)
	}

	return fmt.Sprintf("%x", (h.Sum(nil))), nil
}
