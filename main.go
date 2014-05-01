package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"./api"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var APP_VERSION string
var staticPath = flag.String("static", "../E.A.K/public",
	"The location of the static assets to serve")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	APP_VERSION = getVersion()
	log.Println("Version:", APP_VERSION)

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

func getVersion() string {
	v := run("git", "rev-parse", "HEAD")
	if run("git", "status", "--porcelain") != "" {
		v += "_DEV"
	}
	return v
}

func run(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return strings.Trim(out.String(), " \t\n\r")
}
