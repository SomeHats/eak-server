package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zenazn/goji/web"
)

func currentUserMiddleware(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user, err := getCurrentUser(r)
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

func getCurrentUser(r *http.Request) (User, error) {
	// TODO: use a query that selects the user on some proper metric rather than a default id:
	st := time.Now()
	rows, err := queries.getUser.Query(defaultUser)
	defer rows.Close()
	if err != nil {
		return User{}, err
	}
	log.Println("Got user in", time.Since(st))

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.State, &user.Created, &user.Seen); err != nil {
			return user, err
		} else {
			// Checkin the user, but don't block the request to do so:
			go checkinUser(user.Id)

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
