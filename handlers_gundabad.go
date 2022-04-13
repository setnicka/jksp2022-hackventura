package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const gundabadFinalURL = "/gundabad"

func gundabadRouter() *chi.Mux {
	r := newGameRouter("gundabad")
	r.Get("/", auth(gundabadIndexGet))
	r.Post("/", auth(gundabadIndexPost))
	r.Get(gundabadFinalURL, auth(gundabadInsideGet))
	return r
}

func gundabadIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.Gundabad.Completed {
		http.Redirect(w, r, gundabadFinalURL, http.StatusSeeOther)
		return
	}

	data := getGeneralData("Gundabad", w, r)
	executeTemplate(w, "gundabadIndex", data)
}

func gundabadIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Gundabad.Completed) {
		http.Redirect(w, r, gundabadFinalURL, http.StatusSeeOther)
		return
	}

	team.Gundabad.Tries++
	defer server.state.Save()

	password := r.PostFormValue("password")
	log.Infof("[Gundabad - %s] Trying password '%s'", team.Login, password)

	var was time.Time
	lastTrue := false
	for _, p := range server.state.GetGundabadPasswords() {
		if p.Password == password {
			was = p.From
			lastTrue = true
		} else {
			lastTrue = false
		}
	}

	if lastTrue {
		log.Infof("[Gundabad - %s] Completed", team.Login)
		team.Gundabad.Completed = true
		team.Gundabad.CompletedTime = time.Now()
		http.Redirect(w, r, gundabadFinalURL, http.StatusSeeOther)
	} else if !was.IsZero() {
		d := time.Now().Sub(was)
		if d.Seconds() < 5 {
			setFlashMessage(w, r, messageWarn, "Tohle heslo jsme právě přestali používat, vždyť jste ho teď změnil! Bránu neotevřeme!")
		} else if d.Seconds() > 60 {
			setFlashMessage(w, r, messageWarn, "Tohle je staré heslo, to jsme používali před %d minutami! Bránu neotevřeme!", int(d.Minutes()))
		} else {
			setFlashMessage(w, r, messageWarn, "Tohle je staré heslo, to jsme používali před %d sekundami! Bránu neotevřeme!", int(d.Seconds()))
		}
	} else {
		setFlashMessage(w, r, messageWarn, "Cože?!? To měla být nadávka?")
	}
}

func gundabadInsideGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Gundabad.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Gundabad", w, r)
	executeTemplate(w, "gundabadInside", data)
}
