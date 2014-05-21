package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zenazn/goji/web"
)

func getEventSumHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	types := getEscapedTypes(r)
	if len(types) == 0 {
		http.Error(w, "You must supply types!", http.StatusBadRequest)
	}

	q := fmt.Sprintf(`
		SELECT type, COUNT(*)
		FROM events
		WHERE type IN (%s)
		GROUP BY type`, strings.Join(types, ","))

	st := time.Now()
	rows, err := db.Query(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var k string
		var v int
		if err := rows.Scan(&k, &v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result[k] = v
	}
	log.Println("[sql] Got event aggregate stats in", time.Since(st))

	sendJSON(w, result)
}

func getEventSeriesHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	types := getEscapedTypes(r)
	if len(types) == 0 {
		http.Error(w, "You must supply types!", http.StatusBadRequest)
		return
	}

	q := fmt.Sprintf(`
		SELECT DATE_TRUNC('day', start_time) AS day, type, COUNT(*)
		FROM events
		WHERE type IN (%s)
		GROUP BY day, type
		ORDER BY day`, strings.Join(types, ","))

	st := time.Now()
	rows, err := db.Query(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make(map[string]map[string]int)
	for rows.Next() {
		var t string
		var d time.Time
		var v int

		if err := rows.Scan(&d, &t, &v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := result[t]; !ok {
			result[t] = make(map[string]int)
		}

		result[t][d.Format("2006/01/02")] = v
	}
	log.Println("[sql] Got aggregate series stats in", time.Since(st))

	ensureSeriesFilled(&result)

	sendJSON(w, result)
}

// Feels inefficient... TODO: Make less inefficient!?
func ensureSeriesFilled(series *map[string]map[string]int) {
	for _, ser := range *series {
		for date, _ := range ser {
			for _, sser := range *series {
				if _, ok := sser[date]; !ok {
					sser[date] = 0
				}
			}
		}
	}
}

func getEscapedTypes(r *http.Request) []string {
	typestr := r.URL.Query().Get("types")
	if typestr == "" {
		return []string{}
	}

	types := strings.Split(typestr, ",")
	for i, t := range types {
		types[i] = escape(t, 0)
	}

	return types
}
