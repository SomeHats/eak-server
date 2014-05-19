package api

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/zenazn/goji/web"
)

var db *sql.DB
var defaultUser int
var store *sessions.CookieStore

func Attach(app *web.Mux, conf Config) {
	log.Println("Connecting to Postgres...")
	db = connectPg(conf.Postgres)
	prepareQueries()
	getOrCreateDefaultUser()
	log.Println("Connected")

	store = sessions.NewCookieStore(
		[]byte(conf.Session.Authentication),
		[]byte(conf.Session.Encryption))

	app.Use(currentUserMiddleware)

	app.Get("/api/users/me", getCurrentUserHandler)
	app.Post("/api/users/me/persona", personaLoginHandler)
	app.Post("/api/events", postEventHandler)
	app.Get("/api/events/:id", getEventHandler)
	app.Post("/api/events/:id/checkin", postCheckinHandler)
}

func connectPg(conf PgConfig) *sql.DB {
	conn := fmt.Sprintf(
		"dbname='%s' user='%s' password='%s' host='%s' port='%d' sslmode='%s'",
		conf.Database,
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.SSL,
	)

	var err error
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal("Could not connect to Postgres: ", err)
	}

	db.SetMaxOpenConns(25)

	var ping string
	err = db.QueryRow("SELECT 'ping'").Scan(&ping)
	if err != nil {
		log.Fatal("Could not ping Postgres: ", err)
	}

	if ping != "ping" {
		log.Fatal(`Something weird is going on. "SELECT 'ping'" returned '%v'.`, ping)
	}

	_, err = db.Exec("SET TIMEZONE = 'UTC'")
	if err != nil {
		log.Fatal("Could not set timezone: ", err)
	}

	return db
}

func getOrCreateDefaultUser() {
	// TODO: Remove this function, and store events against actual users
	var id int

	row, err := db.Query("SELECT id FROM users ORDER BY id LIMIT 1")
	defer row.Close()
	if err != nil {
		log.Fatal("Couldn't query for users ", err)
	}

	if row.Next() {
		err := row.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = db.QueryRow(`
			INSERT INTO users (state, created, last_seen)
			VALUES ('implicit', NOW(), NOW())
			RETURNING id
		`).Scan(&id)

		if err != nil {
			log.Fatal("Couldn't create default user ", err)
		}
	}

	log.Println("API default user id:", id)
	defaultUser = id
}
