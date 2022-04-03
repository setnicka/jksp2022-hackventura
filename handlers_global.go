package main

import (
	"net/http"
	"path"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// GeneralData are send to each template
type GeneralData struct {
	Title       string
	User        string
	MessageType flashMessageType
	Message     string
}

func getGeneralData(title string, w http.ResponseWriter, r *http.Request) GeneralData {
	data := GeneralData{
		Title: title,
	}
	if flashMessages := getFlashMessages(w, r); len(flashMessages) > 0 {
		data.MessageType = flashMessages[0].Type
		data.Message = flashMessages[0].Message
		log.Debugf("Flash message '%s'", flashMessages[0].Message)
	}
	return data
}

// router used as base for all targets
func newGameRouter(name string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	staticHandler := http.FileServer(NoListFileSystem{http.Dir(path.Join(staticDir, name))})
	r.Mount("/static/", http.StripPrefix("/static/", staticHandler))
	r.Handle("/favicon.ico", staticHandler)

	r.Get("/start-hry", loginGetHandler)
	r.Post("/start-hry", loginPostHandler)
	r.Get("/konec-hry", logoutHandler)
	return r
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	if checkSession(w, r, true) {
		if server.state.GetTeam(getUser(r)) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}

	data := getGeneralData("Poštovní hackovací proxy", w, r)
	executeTemplate(w, "login", data)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		http.Redirect(w, r, "/start-hry", http.StatusSeeOther)
		return
	}
	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	if ok, team := server.state.TeamLogin(login, password); ok {
		session, _ := server.sessionStore.Get(r, sessionCookieName)
		session.Values["authenticated"] = true
		session.Values["user"] = team.Login
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	setFlashMessage(w, r, messageError, "Nesprávný login")
	http.Redirect(w, r, "/start-hry", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := server.sessionStore.Get(r, sessionCookieName)
	delete(session.Values, "authenticated")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
