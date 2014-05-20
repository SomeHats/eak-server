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
	typestr := r.URL.Query().Get("types")
	if typestr == "" {
		http.Error(w, "You must supply types!", http.StatusBadRequest)
		return
	}

	types := strings.Split(typestr, ",")
	for i, t := range types {
		types[i] = escape(t, 0)
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
