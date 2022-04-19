package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	roklinkaLogin    = "Thurinion"        // Thurin = tajemství, ion = syn
	roklinkaPassword = "elRoNdUvPLNOvous" // musí mít sudý počet znaků
	roklinkaFinalURL = "/imladris"        // sindarský název pro Roklinku
)

func roklinkaRouter() *chi.Mux {
	r := newGameRouter("roklinka")
	r.Get("/", auth(roklinkaIndexGet))
	r.Post("/", auth(roklinkaIndexPost))
	r.Get(roklinkaFinalURL, auth(roklinkaInternalGet))
	return r
}

func b(fail bool) string {
	if fail {
		return "x"
	}
	return "."
}

func roklinkaIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.Roklinka.Completed {
		http.Redirect(w, r, roklinkaFinalURL, http.StatusSeeOther)
		return
	}

	data := getGeneralData("Roklinka", w, r)
	executeTemplate(w, "roklinkaIndex", data)
}

func roklinkaIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Roklinka.Completed) {
		http.Redirect(w, r, roklinkaFinalURL, http.StatusSeeOther)
		return
	}

	team.Roklinka.Tries++
	defer server.state.Save()

	////////////////////////////////////////////////////////////////

	a1 := strings.TrimSpace(r.PostFormValue("a1"))   // první půlka hesla
	a2 := strings.TrimSpace(r.PostFormValue("a2"))   // login bez prvního písmenka
	a3 := strings.TrimSpace(r.PostFormValue("a3"))   // login
	a4 := strings.TrimSpace(r.PostFormValue("a4"))   // prázdné
	a5 := strings.TrimSpace(r.PostFormValue("a5"))   // 2021 (letopočet minulého léta)
	a6 := strings.TrimSpace(r.PostFormValue("a6"))   // druhý znak hesla
	a7 := strings.TrimSpace(r.PostFormValue("a7"))   // password
	a8 := strings.TrimSpace(r.PostFormValue("a8"))   // roklinka (český název tohoto místa)
	a9 := strings.TrimSpace(r.PostFormValue("a9"))   // login pozpátku
	a10 := strings.TrimSpace(r.PostFormValue("a10")) // každý druhý znak hesla

	pwHalf := roklinkaPassword[:len(roklinkaPassword)/2]
	loginWithoutFirst := roklinkaLogin[1:]
	loginReversed := ""
	for i := len(roklinkaLogin) - 1; i >= 0; i-- {
		loginReversed += string(roklinkaLogin[i])
	}
	pwEverySecond := ""
	for i := 0; i < len(roklinkaPassword); i += 2 {
		pwEverySecond += string(roklinkaPassword[i])
	}

	c1 := a1 != pwHalf
	c2 := a2 != loginWithoutFirst
	c3 := a3 != roklinkaLogin
	c4 := len(a4) != 0
	c5 := a5 != "2021"
	c6 := len(a6) < 1 || a6[0] != roklinkaPassword[1]
	c7 := a7 != roklinkaPassword
	c8 := strings.ToLower(a8) != "roklinka"
	c9 := a9 != loginReversed
	c10 := a10 != pwEverySecond

	tests := b(c1) + b(c2) + b(c3) + b(c4) + b(c5) + b(c6) + b(c7) + b(c8) + b(c9) + b(c10)
	log.Infof("[Roklinka - %s] Trying login '%s' and password '%s': %s", team.Login, a3, a7, tests)
	fail := c1 || c2 || c3 || c4 || c5 || c6 || c7 || c8 || c9 || c10

	if fail {
		// setFlashMessage(w, r, messageError, " ") // "Nesprávné údaje"
		testsFrontendOrder := b(c4) + b(c2) + b(c8) + b(c7) + b(c10) + b(c3) + b(c5) + b(c1) + b(c9) + b(c6)
		setFlashMessage(w, r, messageError, "  "+testsFrontendOrder) // "Nesprávné údaje: ...x..."
		return
	}

	log.Infof("[Roklinka - %s] Completed", team.Login)
	// Everything completed
	team.Roklinka.Completed = true
	team.Roklinka.CompletedTime = time.Now()
	http.Redirect(w, r, roklinkaFinalURL, http.StatusSeeOther)
}

func roklinkaInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Roklinka.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Roklinka", w, r)
	executeTemplate(w, "roklinkaInternal", data)
}
