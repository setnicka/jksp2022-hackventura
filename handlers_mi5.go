package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	mi5Login    = "EricForbes"
	mi5Password = "Pass123!"

	mi5AssetsDir = "mi5assets"
)

func mi5Router() *chi.Mux {
	r := newGameRouter("mi5")
	r.Get("/", auth(mi5IndexGet))
	r.Post("/", auth(mi5IndexPost))
	r.Get("/intranet", auth(mi5InternalGet))
	r.Mount("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(mi5AssetsDir))))
	return r
}

func mi5IndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.MI5.Completed {
		http.Redirect(w, r, "/intranet", http.StatusSeeOther)
		return
	}

	data := getGeneralData("MI5", w, r)
	executeTemplate(w, "mi5Index", data)
}
func mi5IndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	team.MI5.Tries++
	defer server.state.Save()

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[MI5 - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if login != mi5Login || password != mi5Password {
		setFlashMessage(w, r, messageError, "Invalid credentials")
		return
	}

	log.Infof("[MI5 - %s] Completed", team.Login)
	// Everything completed
	team.MI5.Completed = true
	team.MI5.CompletedTime = time.Now()
	http.Redirect(w, r, "/intranet", http.StatusSeeOther)
}

func mi5InternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.MI5.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("MI5", w, r)
	executeTemplate(w, "mi5Internal", data)
}
