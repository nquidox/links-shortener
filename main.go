package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "")
}

func main() {
	config, err := loadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	dbFile := config.DBSource
	db, err := sqlx.Open(config.DBDriver, dbFile)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	if !exists(dbFile) {
		db.MustExec(`CREATE TABLE IF NOT EXISTS links (id INTEGER, long_link TEXT, short_link TEXT, PRIMARY KEY ("id" AUTOINCREMENT))`)
	}

	shortenHandler := func(c echo.Context) error {
		userLink := c.FormValue("user_link")

		longLink, shortLink, _ := isPresent(db, userLink)

		if longLink == userLink {
			return c.Render(http.StatusOK, "shortenedLink", shortLink)
		} else if longLink != userLink {
			shortLink, _ := insertNewLink(db, userLink)
			return c.Render(http.StatusOK, "shortenedLink", shortLink)
		}
		return nil
	}

	redirectHandler := func(c echo.Context) error {
		userShortLink := strings.Trim(c.Request().URL.String(), "/")

		findLink, _ := findShortLink(db, userShortLink)

		if findLink != "" {
			switch {
			case strings.Contains(findLink, "http://") || strings.Contains(findLink, "https://"):
				return c.Redirect(http.StatusMovedPermanently, findLink)

			default:
				return c.Redirect(http.StatusMovedPermanently, "https://"+findLink)

			}
		} else {
			return c.Render(http.StatusOK, "notFound", "")
		}
	}

	e := echo.New()
	e.Renderer = &Template{templates: template.Must(template.ParseGlob("templates/*.html"))}
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/media/globe.svg")
	e.GET("/", index)
	e.POST("/", shortenHandler)
	e.GET("/:", redirectHandler)

	log.Fatal(e.Start(config.ServerAddress))
}
