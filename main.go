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

var staticPath = flag.String("static", "../E.A.K/public",
	"The location of the static assets to serve")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	conf := loadConfig()
	if conf.ApiEnabled {
		apiHandler := web.New()
		goji.Handle("/api/*", apiHandler)
		api.Attach(apiHandler, APP_VERSION, conf)
	}

	goji.Get("/*", static(*staticPath))

	goji.Serve()
	log.Println("Finished")
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 :(", http.StatusNotFound)
}
