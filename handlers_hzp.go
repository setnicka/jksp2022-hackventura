package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	hzpLogin    = "75841"
	hzpPassword = "baltazar"
	hzpName     = "Peckoslav Orgorád" // display in frontend
	hzpFinalURL = "/moje-hzp"
)

func hzpRouter() *chi.Mux {
	r := newGameRouter("hzp")
	r.Get("/", auth(hzpIndexGet))
	r.Post("/", auth(hzpIndexPost))
	r.Get(hzpFinalURL, auth(hzpIntranetGet))
	return r
}

type hzpIndexData struct {
	GeneralData
	Completed bool
	Name      string
}

func hzpIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))

	data := hzpIndexData{
		GeneralData: getGeneralData("HZP – Hobití Zdravotní Pojišťovna", w, r),
		Completed:   team.HZP.Completed,
		Name:        hzpName,
	}
	executeTemplate(w, "hzpIndex", data)
}

func hzpIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.HZP.Completed) {
		http.Redirect(w, r, hzpFinalURL, http.StatusSeeOther)
		return
	}

	team.HZP.Tries++
	defer server.state.Save()

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[HZP - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if login == hzpLogin && strings.ToLower(password) == hzpPassword {
		log.Infof("[HZP - %s] Completed", team.Login)
		// Everything completed
		team.HZP.Completed = true
		team.HZP.CompletedTime = time.Now()
		setFlashMessage(w, r, messageOk, "Přihlášení bylo úspěšné, vítejte v systému Moje HZP")
		http.Redirect(w, r, hzpFinalURL, http.StatusSeeOther)
	} else {
		setFlashMessage(w, r, messageError, "Nesprávné číslo průkazu nebo heslo")
	}
}

func hzpIntranetGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.HZP.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := hzpIndexData{
		GeneralData: getGeneralData("HZP – Hobití Zdravotní Pojišťovna", w, r),
		Completed:   team.HZP.Completed,
		Name:        hzpName,
	}

	executeTemplate(w, "hzpIntranet", data)
}
