package main

import (
	"flag"
	"html/template"
	"net/http"
	"strings"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"

	"github.com/setnicka/jksp2022-hackventura/state"
)

const (
	sessionCookieSecret = "Aingairaec6pheNg5wieThae7moofe"
	sessionMaxAge       = 3600 * 24
	sessionCookieName   = "hack_systems"
	templatesDir        = "templates"
	staticDir           = "static"

	orgLogin    = "ksp"
	orgPassword = "ksp" // TODO: change before the game
)

// Server represents the whole system with its state
type Server struct {
	sessionStore sessions.Store
	templates    *template.Template
	state        *state.State
}

type subdomains map[string]*chi.Mux

var (
	listen = flag.String("listen", ":8080", "Listen address")
)

// Global singleton
var server *Server

////////////////////////////////////////////////////////////////////////////////

type targetS struct {
	Code   string
	Name   string
	URL    string
	Router func() *chi.Mux
}

// const baseDomain = "setnicka.cz:8080"
const baseDomain = "localhost:8080"

var targets = []targetS{
	{Code: "moria", Name: "Mórie", URL: "moria." + baseDomain, Router: moriaRouter},
	{Code: "gundabad", Name: "Gundabad", URL: "gundabad." + baseDomain, Router: gundabadRouter},
	{Code: "roklinka", Name: "Roklinka", URL: "roklinka." + baseDomain, Router: roklinkaRouter},
	{Code: "erebor", Name: "Erebor", URL: "erebor." + baseDomain, Router: ereborRouter},
	{Code: "dunharrow", Name: "Dunharrow", URL: "dunharrow." + baseDomain, Router: dunharrowRouter},
	{Code: "hzp", Name: "HZP", URL: "hzp." + baseDomain, Router: hzpRouter},
	{Code: "village", Name: "Kamenná vesnice", URL: "village." + baseDomain, Router: villageRouter},
}

////////////////////////////////////////////////////////////////////////////////

func (sub subdomains) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domainParts := strings.Split(r.Host, ".")

	if mux := sub[domainParts[0]]; mux != nil {
		// Let the appropriate mux serve the request
		mux.ServeHTTP(w, r)
	} else {
		// Handle 404
		http.Error(w, "Subdomain not found", 404)
	}
}

func main() {
	cookieStore := sessions.NewCookieStore([]byte(sessionCookieSecret))
	cookieStore.MaxAge(sessionMaxAge)
	//cookieStore.Options.Domain = ".fuf.me"

	server = &Server{
		sessionStore: cookieStore,
		state:        state.Init(),
	}

	server.start()
}

func (s *Server) start() {
	log.Info("Starting server...")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	//r.Use(middleware.Logger) // used in subrouters independently
	r.Use(middleware.Recoverer)

	// Routers for subdomains:
	sub := subdomains{
		"org": orgRouter(),
	}
	for _, target := range targets {
		sub[target.Code] = target.Router()
	}

	if _, err := server.getTemplates(); err != nil {
		log.Errorf("Cannot load templates: %v", err)
		return
	}

	r.Mount("/", sub)
	log.Infof("Server started at %s", *listen)
	if err := http.ListenAndServe(*listen, r); err != nil {
		log.Errorf("SERVER ERROR: %+v", err)
	}
}
