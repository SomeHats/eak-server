package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zenazn/goji/web"
)

const year = 60 * 60 * 24 * 365

func currentUserMiddleware(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user, err := getCurrentUser(r, w)
		if err != nil {
			log.Println("[currentUserMiddleware] Error: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c.Env["user"] = user
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func getEnvUser(env map[string]interface{}) User {
	user, ok := env["user"].(User)
	if ok {
		return user
	} else {
		return User{}
	}
}

func getCurrentUserHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	sendJSON(w, c.Env["user"])
}

func getCurrentUser(r *http.Request, w http.ResponseWriter) (User, error) {
	session, err := store.Get(r, "eak-session")
	if err != nil {
		return User{}, err
	}
	session.Options.MaxAge = year

	var rows *sql.Rows

	id, ok := session.Values["user-id"]
	if ok {
		st := time.Now()
		rows, err = queries.getUser.Query(id)
		defer rows.Close()
		if err != nil {
			return User{}, err
		}
		log.Println("Got user in", time.Since(st))
	} else {
		st := time.Now()
		rows, err = queries.createImplicitUser.Query()
		defer rows.Close()
		if err != nil {
			return User{}, err
		}
		log.Println("Created user in", time.Since(st))
	}

	userPtr, err := scanUser(rows)
	if err != nil {
		return User{}, err
	}
	if userPtr == nil {
		return User{}, fmt.Errorf("Could not find current user :(")
	}
	user := *userPtr

	// Checkin the user, but don't block the request to do so:
	go checkinUser(user.Id)

	// Save session:
	session.Values["user-id"] = user.Id
	err = session.Save(r, w)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func scanUser(rows *sql.Rows) (*User, error) {
	user := User{}
	email := sql.NullString{}
	for rows.Next() {
		err := rows.Scan(&user.Id, &user.State, &email, &user.Created, &user.Seen)
		if err != nil {
			return nil, err
		}

		user.Email = email.String
		return &user, nil
	}

	return nil, nil
}

func checkinUser(id int) {
	st := time.Now()
	_, err := queries.updateUser.Exec(id)
	if err != nil {
		log.Printf("[non-blocking] Failed to update user %d: %v", id, err)
	} else {
		log.Printf("[non-blocking] Updated user %d in %v", id, time.Since(st))
	}
}
