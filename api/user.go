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

func getCurrentUserHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	sendJSON(w, c.Env["user"])
}

func getCurrentUser(r *http.Request, w http.ResponseWriter) (User, error) {
	session, err := store.Get(r, "eak-session")
	if err != nil {
		return User{}, err
	}

	var rows *sql.Rows

	session.Options.MaxAge = year
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

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.State, &user.Created, &user.Seen); err != nil {
			return user, err
		} else {
			// Checkin the user, but don't block the request to do so:
			go checkinUser(user.Id)

			// Save session:
			session.Values["user-id"] = user.Id
			err := session.Save(r, w)
			if err != nil {
				return User{}, err
			}

			return user, nil
		}
	}

	return User{}, fmt.Errorf("Could not find current user :(")
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
