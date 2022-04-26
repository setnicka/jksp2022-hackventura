package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

type moriaQuestion struct {
	Question string
	Answer   string
}

var moriaQuestions = []moriaQuestion{
	{"Řekni sindarilsky 'přítel' a vejdi", "Mellon"},
	{"Vyslov jméno sekery slavného Gimliho trpaslíka", "drtihlavka"},
	{"Jmenuj nápoj jež nejoblíbenější Elrondův jest", "miruvor"},
	{"Jaká je průměrná odpovědní doba enta?", "semestr"},
	{"Jaký je Šmáčkův roční obrat v pytlech zlata", "257"},
	{"Kolik let je hospodskému z Hůrky", "37.856"},
}

var moriaFinalURL = "/" + url.PathEscape("khazad-dûm")

func moriaRouter() *chi.Mux {
	r := newGameRouter("moria")
	r.Get("/", auth(moriaIndexGet))
	r.Post("/", auth(moriaIndexPost))
	r.Get("/khazad-dûm", auth(moriaInternalGet))
	return r
}

type moriaIndexData struct {
	GeneralData
	Completed bool
	Image     string
	Question  string
}

func moriaIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Moria.Completed) {
		http.Redirect(w, r, moriaFinalURL, http.StatusSeeOther)
		return
	}
	if team.Moria.RightAnswers >= len(moriaQuestions) {
		http.Error(w, "Chybí další otázka, běž za orgy", 500)
		return
	}

	data := moriaIndexData{
		GeneralData: getGeneralData("Brána do Mórie", w, r),
		Completed:   false,
		Image:       fmt.Sprintf("durin_door%d.png", team.Moria.RightAnswers),
		Question:    moriaQuestions[team.Moria.RightAnswers].Question,
	}

	getGeneralData("Brána do Mórie", w, r)
	executeTemplate(w, "moriaIndex", data)
}

func moriaIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && (team.Moria.Completed) {
		http.Redirect(w, r, moriaFinalURL, http.StatusSeeOther)
		return
	}
	if team.Moria.RightAnswers >= len(moriaQuestions) {
		http.Error(w, "Chybí další otázka, běž za orgy", 500)
		return
	}

	team.Moria.Tries++
	defer server.state.Save()

	answer := r.PostFormValue("answer")
	log.Infof("[Moria - %s] Trying answer '%s' to question %d", team.Login, answer, team.Moria.RightAnswers)
	// normalization
	answer = strings.ToLower(strings.Replace(answer, ",", ".", 1))

	if answer == strings.ToLower(moriaQuestions[team.Moria.RightAnswers].Answer) {
		log.Infof("[Moria - %s] Right answer to question %d", team.Login, team.Moria.RightAnswers)
		team.Moria.RightAnswers++

		if team.Moria.RightAnswers == len(moriaQuestions) {
			log.Infof("[Moria - %s] Completed", team.Login)
			team.Moria.Completed = true
			team.Moria.CompletedTime = time.Now()
			http.Redirect(w, r, moriaFinalURL, http.StatusSeeOther)
		} else {
			setFlashMessage(w, r, messageOk, "Brána zazářila a další z vrchních hvězd se rozzářila zlatavým svitem")
		}
	} else {
		setFlashMessage(w, r, messageError, "Vyřkli jste vaši odpověď. S bránou to však ani nehlo…")
	}
}

func moriaInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Moria.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := moriaIndexData{
		GeneralData: getGeneralData("Uvnitř Khazad-dûm", w, r),
		Completed:   true,
		Image:       "durin_door_fin.png",
	}

	executeTemplate(w, "moriaIndex", data)
}
