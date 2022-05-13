package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	villageFinalURL = "/village"
	villagePassword = "hobyte"
)

func villageRouter() *chi.Mux {
	r := newGameRouter("village")
	r.Get("/", auth(villageIndexGet))
	r.Post("/", auth(villageIndexPost))
	r.Get(villageFinalURL, auth(villageInsideGet))
	return r
}

func villageIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.Village.Completed {
		http.Redirect(w, r, villageFinalURL, http.StatusSeeOther)
		return
	}

	data := getGeneralData("Village", w, r)
	executeTemplate(w, "villageIndex", data)
}

func villageIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Village.Completed) {
		http.Redirect(w, r, villageFinalURL, http.StatusSeeOther)
		return
	}

	team.Village.Tries++
	defer server.state.Save()

	password := r.PostFormValue("password")
	log.Infof("[Village - %s] password '%s'", team.Login, password)


	if password == villagePassword {
		log.Infof("[Village - %s] Completed", team.Login)
		team.Village.Completed = true
		team.Village.CompletedTime = time.Now()
		http.Redirect(w, r, villageFinalURL, http.StatusSeeOther)
	} else {
		setFlashMessage(w, r, messageWarn, "")
	}
}

func villageInsideGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Village.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Village", w, r)
	executeTemplate(w, "villageInside", data)
}
