package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/qwsnxnjene/telegram-bot/storage"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("[New]: can't open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("[New]: can't connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, p *storage.Page) error {
	q := `INSERT INTO pages (url, user_name, title, date) VALUES (:url, :username, :title, :date)`

	if _, err := s.db.ExecContext(ctx, q,
		sql.Named("url", p.URL),
		sql.Named("username", p.UserName),
		sql.Named("title", p.Title),
		sql.Named("date", p.Date)); err != nil {
		return fmt.Errorf("[Save]: can't save page: %w", err)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, userName string) (*storage.Page, error) {
	q := `SELECT url, title, date FROM pages WHERE user_name = :username ORDER BY RANDOM() LIMIT 1`

	var url, title, date string

	err := s.db.QueryRowContext(ctx, q, sql.Named("username", userName)).Scan(&url, &title, &date)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("[PickRandom]: can't pick random page: %w", err)
	}

	return &storage.Page{
		URL:      url,
		UserName: userName,
		Title:    title,
		Date:     date,
	}, nil
}

func (s *Storage) Remove(ctx context.Context, page *storage.Page) error {
	q := `DELETE FROM pages WHERE url = :url AND user_name = :username`

	if _, err := s.db.ExecContext(ctx, q, sql.Named("url", page.URL), sql.Named("username", page.UserName)); err != nil {
		return fmt.Errorf("[Remove]: can't delete from page: %w", err)
	}

	return nil
}

func (s *Storage) IsExists(ctx context.Context, page *storage.Page) (bool, error) {
	q := `SELECT COUNT(*) FROM pages WHERE url = :url AND user_name = :username`

	var count int

	if err := s.db.QueryRowContext(ctx, q, sql.Named("url", page.URL), sql.Named("username", page.UserName)).Scan(&count); err != nil {
		return false, fmt.Errorf("[IsExists]: can't check if page exists: %w", err)
	}

	return count > 0, nil
}

func (s *Storage) PickLastFive(ctx context.Context, username string) ([]*storage.Page, error) {
	limit := 5

	q := `SELECT url, title, date FROM pages WHERE user_name = :username ORDER BY date DESC LIMIT :limit`
	rows, err := s.db.QueryContext(ctx, q, sql.Named("username", username), sql.Named("limit", limit))
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("[PickLastFive]: can't pick last five pages: %w", err)
	}

	var pages []*storage.Page

	for rows.Next() {
		var url, title, date string

		if err := rows.Scan(&url, &title, &date); err == sql.ErrNoRows {
			return nil, err
		} else if err != nil {
			return nil, fmt.Errorf("[PickLastTen]: can't pick last ten pages: %w", err)
		}

		pages = append(pages, &storage.Page{URL: url, UserName: username, Title: title, Date: date})
	}

	return pages, nil
}

func (s *Storage) Init(ctx context.Context) error {
	check := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='pages'`
	var count int

	if err := s.db.QueryRowContext(ctx, check).Scan(&count); err != nil {
		return fmt.Errorf("[Init]: can't check if table exists: %w", err)
	}

	if count != 0 {
		return nil
	}

	q := `CREATE TABLE IF NOT EXISTS pages 
	(url TEXT NOT NULL DEFAULT "",
	 user_name TEXT NOT NULL DEFAULT "",
	 title TEXT, 
	 date CHAR(8) NOT NULL DEFAULT "19700101");
	 
	 CREATE INDEX user_date ON pages(user_name, date);`

	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("[Init]: can't create table: %w", err)
	}

	return nil
}
