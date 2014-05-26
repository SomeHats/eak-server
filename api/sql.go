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
	eventSession2Child *sql.Stmt
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
		SET last_seen = STATEMENT_TIMESTAMP()
		WHERE id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.updateUser = q

	q, err = db.Prepare(`
		UPDATE users
		SET email = $2, state = 'active'
		WHERE id = $1
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.updateUserEmail = q

	q, err = db.Prepare(`
		INSERT INTO users (state, created, last_seen)
		VALUES ('implicit', STATEMENT_TIMESTAMP(), STATEMENT_TIMESTAMP())
		RETURNING id, state, email, created, last_seen
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.createImplicitUser = q

	q, err = db.Prepare(`
		INSERT INTO events (user_id, parent_id, type, version, start_time, duration, event_data)
		VALUES ($1, $2, $3, $4, STATEMENT_TIMESTAMP(), NULL, $5)
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

	q, err = db.Prepare(`
		SELECT DISTINCT events.id, events.user_id, events.parent_id, events.type,
			events.version, events.start_time, events.duration, '{}' AS event_data
		FROM events, (
			SELECT events.id
			FROM events, (
				SELECT id
				FROM events
				WHERE type = 'session'
				ORDER BY id DESC
				OFFSET $1
				LIMIT $2
			) AS tmp
			WHERE events.id = tmp.id OR events.parent_id = tmp.id
		) AS tmp
		WHERE events.id = tmp.id OR events.parent_id = tmp.id
		ORDER BY id
	`)
	if err != nil {
		log.Fatal(err)
	}
	queries.eventSession2Child = q
}
