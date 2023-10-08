package main

import (
	"github.com/jmoiron/sqlx"
	"math/rand"
	"os"
)

func randomString(size int) string {
	const lettersBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, size)
	for i := range b {
		b[i] = lettersBytes[rand.Intn(len(lettersBytes))]
	}
	return string(b)
}

func exists(path string) bool {
	file, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !file.IsDir()
}

func insertNewLink(db *sqlx.DB, longLink string) (string, error) {
	shortLink := randomString(6)
	tx := db.MustBegin()
	tx.MustExec(`INSERT INTO links (long_link, short_link) VALUES ($1, $2)`, longLink, shortLink)
	err := tx.Commit()
	if err != nil {
		return "", err
	}
	return shortLink, nil
}

func findShortLink(db *sqlx.DB, shortLink string) (string, error) {
	var longLink string
	err := db.Get(&longLink, `SELECT long_link FROM links WHERE short_link=$1`, shortLink)
	if err != nil {
		return "", err
	}
	return longLink, err
}

func isPresent(db *sqlx.DB, longLink string) (string, string, error) {
	type Links struct {
		LongLink  string `db:"long_link"`
		ShortLink string `db:"short_link"`
	}

	var some Links
	err := db.Get(&some, `SELECT long_link, short_link FROM links WHERE long_link=$1`, longLink)
	if err != nil {
		return "", "", err
	}

	return some.LongLink, some.ShortLink, nil
}
