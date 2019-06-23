package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/JRaspass/code-golf/routes"
	brotli "github.com/cv-library/negroni-brotli"
	"github.com/urfave/negroni"
)

type err500 struct{}

func (*err500) FormatPanicError(w http.ResponseWriter, r *http.Request, _ *negroni.PanicInformation) {
	http.Error(w, "500: It's Dead, Jim.", 500)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	logger := negroni.NewLogger()
	logger.ALogger = log.New(os.Stdout, "", 0)
	logger.SetFormat("{{.StartTime}} {{.Status}} {{.Method}} {{.Request.URL}} {{.Request.UserAgent}}")

	recovery := negroni.NewRecovery()
	recovery.Formatter = &err500{}
	recovery.Logger = log.New(os.Stderr, "<1>", 0)

	server := &http.Server{
		Addr: ":1337",
		Handler: negroni.New(
			negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				if r.Host == "ng.code-golf.io" {
					url, _ := url.Parse("http://localhost:1337")

					r.URL.Host = url.Host
					r.URL.Scheme = url.Scheme
					r.Host = url.Host

					httputil.NewSingleHostReverseProxy(url).ServeHTTP(w, r)
				} else {
					next(w, r)
				}
			}),
			logger,
			brotli.New(5),
			recovery,
			negroni.Wrap(routes.Router),
		),
	}

	go func() {
		ticker := time.NewTicker(time.Minute)

		for {
			<-ticker.C
			routes.Ideas()
			routes.Stars()
		}
	}()

	println("Listening…")

	panic(server.ListenAndServe())
}
