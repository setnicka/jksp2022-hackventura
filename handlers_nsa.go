package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	nsaLogin    = "patterson"
	nsaPassword = "laoo,rpe" // "password" on Dvorak keyboard
)

func nsaRouter() *chi.Mux {
	r := newGameRouter("nsa")
	r.Get("/", auth(nsaIndexGet))
	r.Post("/", auth(nsaIndexPost))
	r.Get("/intranet", auth(nsaIntranetGet))
	return r
}

func nsaIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.NSA.Completed {
		http.Redirect(w, r, "/intranet", http.StatusSeeOther)
		return
	}

	data := getGeneralData("NSA", w, r)
	executeTemplate(w, "nsaIndex", data)
}

func nsaIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	team.NSA.Tries++
	defer server.state.Save()

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[NSA - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if strings.ToLower(login) == nsaLogin && password == nsaPassword {
		log.Infof("[NSA - %s] Completed", team.Login)
		// Everything completed
		team.NSA.Completed = true
		team.NSA.CompletedTime = time.Now()
		http.Redirect(w, r, "/intranet", http.StatusSeeOther)
	} else {
		setFlashMessage(w, r, messageError, "Incorrect login or password, check your spelling")
	}
}

func nsaIntranetGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.NSA.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("NSA", w, r)
	executeTemplate(w, "nsaIntranet", data)
}
