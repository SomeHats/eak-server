package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func sendJSON(w http.ResponseWriter, js interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(js)
	if err != nil {
		log.Println("JSON error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(b)
}

func parseReqJSON(r *http.Request, d interface{}) (error, int) {
	if r.Body == nil {
		return fmt.Errorf("You must supply a body"), http.StatusBadRequest
	}

	dcd := json.NewDecoder(r.Body)
	if err := dcd.Decode(d); err != nil {
		return fmt.Errorf("Could not decode JSON: %v", err), http.StatusBadRequest
	}

	return nil, http.StatusOK
}

func escape(str string, e int) string {
	escaper := strings.Map(func(r rune) rune { return escMap[r] }, fmt.Sprintf("$%d$", e))

	if strings.Contains(str, escaper) {
		return escape(str, e+1)
	} else {
		return escaper + str + escaper
	}
}

var escMap = map[rune]rune{
	'0': 'a',
	'1': 'b',
	'2': 'c',
	'3': 'd',
	'4': 'e',
	'5': 'f',
	'6': 'g',
	'7': 'h',
	'8': 'i',
	'9': 'j',
	'$': '$',
}
