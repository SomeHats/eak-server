package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"./api"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	conf := loadConfig()
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
	http.Error(w, "404 :(", http.StatusNotFound)
}
