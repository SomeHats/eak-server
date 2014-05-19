package api

import (
	"database/sql"
	"log"
)

var queries struct {
	getUser            *sql.Stmt
	getUserByEmail     *sql.Stmt
	updateUser         *sql.Stmt
	updateUserEmail    *sql.Stmt
	createImplicitUser *sql.Stmt
	createEvent        *sql.Stmt
	getEvent           *sql.Stmt
	checkin            *sql.Stmt
}

func prepareQueries() {
	q, err := db.Prepare(`
		SELECT id, state, email, created, last_seen
		FROM users
		WHERE id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.getUser = q

	q, err = db.Prepare(`
		SELECT id, state, email, created, last_seen
		FROM users
		WHERE email = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.getUserByEmail = q

	q, err = db.Prepare(`
		UPDATE users
		SET last_seen = NOW()
		WHERE id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.updateUser = q

	q, err = db.Prepare(`
		UPDATE users
		SET email = $2
		WHERE id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.updateUserEmail = q

	q, err = db.Prepare(`
		INSERT INTO users (state, created, last_seen)
		VALUES ('implicit', NOW(), NOW())
		RETURNING id, state, email, created, last_seen
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.createImplicitUser = q

	q, err = db.Prepare(`
		INSERT INTO events (user_id, parent_id, type, version, start_time, duration, event_data)
		VALUES ($1, $2, $3, $4, NOW(), NULL, $5)
		RETURNING id, user_id, parent_id, type, version, start_time, duration, event_data
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.createEvent = q

	q, err = db.Prepare(`
		SELECT id, user_id, parent_id, type, version, start_time, duration, event_data
		FROM events
		WHERE id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.getEvent = q

	q, err = db.Prepare(`
		UPDATE events
		SET duration = $2
		WHERE id = $1
		RETURNING id, user_id, parent_id, type, version, start_time, duration, event_data
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.checkin = q
}
