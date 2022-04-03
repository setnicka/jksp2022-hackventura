package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	bisLogin    = "reditel"
	bisPassword = "mujmalygripen"
)

func bisRouter() *chi.Mux {
	r := newGameRouter("bis")
	r.Get("/", auth(bisIndexGet))
	r.Post("/", auth(bisIndexPost))
	r.Get("/tajne", auth(bisInternalGet))
	return r
}

func bisIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.BIS.Completed {
		http.Redirect(w, r, "/tajne", http.StatusSeeOther)
		return
	}

	data := getGeneralData("BIS", w, r)
	executeTemplate(w, "bisIndex", data)
}

func bisIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	team.BIS.Tries++
	defer server.state.Save()

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[BIS - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if login != bisLogin || password != bisPassword {
		setFlashMessage(w, r, messageError, "Nesprávný login, zkuste to znovu nebo kontaktujte podporu")
		return
	}

	log.Infof("[BIS - %s] Completed", team.Login)
	// Everything completed
	team.BIS.Completed = true
	team.BIS.CompletedTime = time.Now()
	http.Redirect(w, r, "/tajne", http.StatusSeeOther)
}

func bisInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.BIS.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("BIS", w, r)
	data.MessageType = "success"
	data.Message = "Přístup povolen"
	executeTemplate(w, "bisIndex", data)
}
