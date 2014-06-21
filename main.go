package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"
	"strings"

	"./api"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	conf := loadConfig()
	redirects(conf.Redirects)

	if conf.ApiEnabled {
		apiHandler := web.New()
		goji.Handle("/api/*", apiHandler)
		api.Attach(apiHandler, conf)
	}

	for url, path := range conf.Static {
		goji.Get(url+"*", static(path, url))
	}

	goji.Serve()
	log.Println("Finished")
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func redirects(redirs map[string]string) {
	for from, to := range redirs {
		addRedirect(from, to)
	}
}

func addRedirect(from, to string) {
	goji.Get(from, func(c web.C, w http.ResponseWriter, r *http.Request) {
		t := to
		for key, val := range c.URLParams {
			t = strings.Replace(t, ":"+key, val, -1)
		}

		w.Header().Set("Location", t)
		http.Error(w, "Redirected to "+t, http.StatusMovedPermanently)
	})
}
