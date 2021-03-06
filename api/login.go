package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/zenazn/goji/web"
)

func isAllowedAudience(audience string) bool {
	return audience == "http://eraseallkittens.com" || audience == "http://localhost:3000"
}

func personaLoginHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Assertion string
	}

	err, status := parseReqJSON(r, &body)
	if err != nil {
		http.Error(w, "Error parsing request", status)
		return
	}

	audience := r.Header.Get("x-origin")
	if !isAllowedAudience(audience) {
		http.Error(w, "Audience not allowed: "+audience, http.StatusBadRequest)
		return
	}

	email, err := personaLogin(body.Assertion, audience)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentUser := getEnvUser(c.Env)
	user, err := getOrAssociateUserByEmail(email, currentUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	session, err := store.Get(r, "eak-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = year

	session.Values["user-id"] = user.Id
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, user)
}

func personaLogin(assertion, audience string) (string, error) {
	v := url.Values{}
	v.Set("assertion", assertion)
	v.Set("audience", audience)

	st := time.Now()
	res, err := http.PostForm("https://verifier.login.persona.org/verify", v)
	if err != nil {
		return "", err
	}

	var response struct {
		Status string
		Email  string
	}

	dcd := json.NewDecoder(res.Body)
	if err := dcd.Decode(&response); err != nil {
		return "", fmt.Errorf("Could not decode persona response JSON: %v", err)
	}
	log.Println("[http] Verified persona assertion in", time.Since(st))

	if response.Status != "okay" {
		return "", fmt.Errorf("Couldn't log you in: %#v", response)
	}

	return response.Email, nil
}

func getOrAssociateUserByEmail(email string, currentUser User) (User, error) {
	st := time.Now()
	rows, err := queries.getUserByEmail.Query(email)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	user, err := scanUser(rows)
	if err != nil {
		return User{}, err
	}
	log.Println("[sql] Tried to fetch user by email in", time.Since(st))

	if user == nil {
		if err := assocEmail(email, currentUser); err != nil {
			return User{}, err
		}

		currentUser.Email = email
		return currentUser, nil
	} else {
		return *user, nil
	}
}

func assocEmail(email string, user User) error {
	st := time.Now()
	_, err := queries.updateUserEmail.Exec(user.Id, email)
	log.Println("[sql] Updated user email in", time.Since(st))
	if err != nil {
		return err
	} else {
		return nil
	}
}
