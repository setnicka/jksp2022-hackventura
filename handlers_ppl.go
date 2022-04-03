package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

func pplRouter() *chi.Mux {
	r := newGameRouter("ppl")
	r.Get("/", auth(pplIndexGet))
	r.Get("/internal", auth(pplInternalGet))
	r.HandleFunc("/tracking", auth(pplTrackingGet))
	return r
}

type pplTrackingData struct {
	Code     string
	City     string
	ZipCode  string
	Customer string
	Weight   string
	Location string
}

var packages = []pplTrackingData{
	{"55604070459", "Horní Sytová", "512 41", "Marťanská kolonizační a.s.", "0,1 kg", "Balíček byl pohozen kolem barelu u budovy."},
	{"16145962924", "Horní Sytová", "512 41", "Elektrárna Dukovany", "9,5 kg", "Balíček byl dodán za sloup železničního mostu, jak bylo domluveno."},
	{"35015671123", "Horní Sytová", "512 41", "Pražský urychlovač částic", "10 tun", "Balíček je v hajzlu. První patro, páni, levá kabinka."},
	{"44467405130", "Horní Sytová", "512 41", "Pan F, ČP", "0,19 kg", "Balíček byl úspěšně doručen. Access Granted."},
}

type pplTrackingPageData struct {
	GeneralData
	PplTrackingData pplTrackingData
}

func pplIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.PPL.Completed {
		http.Redirect(w, r, "/internal", http.StatusSeeOther)
		return
	}

	data := getGeneralData("PPL", w, r)
	executeTemplate(w, "pplIndex", data)
}

func pplInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.PPL.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("PPL", w, r)
	executeTemplate(w, "pplInternal", data)
}

func pplTrackingGet(w http.ResponseWriter, r *http.Request) {
	data := pplTrackingPageData{GeneralData: getGeneralData("PPL", w, r)}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if team != nil && team.PPL.Completed {
		http.Redirect(w, r, "/internal", http.StatusSeeOther)
		return
	}

	team.PPL.Tries++
	defer server.state.Save()

	code := r.URL.Query().Get("zasilka_id")
	log.Infof("[PPL - %s] Trying package id: %s", team.Login, code)

	for i, pack := range packages {
		if code == pack.Code {
			if team.PPL.PackagesTracked == i {
				team.PPL.PackagesTracked = i + 1
				log.Infof("[PPL - %s] Tracked package %d", team.Login, i+1)
			}
			if team.PPL.PackagesTracked >= i+1 {
				data.PplTrackingData = pack
			} else {
				data.Message = "Zásilka nebyla nalezena kvůli chybě v časoprostorovém kontinuu."
				log.Infof("[PPL - %s] Tried to track package %d without discovering package %d", team.Login, i+1, team.PPL.PackagesTracked+1)
			}
		}
	}

	if team.PPL.PackagesTracked == len(packages) && !team.PPL.Completed {
		data.Message = "Access Granted"
		team.PPL.Completed = true
		team.PPL.CompletedTime = time.Now()
	}

	if data.PplTrackingData.Code != "" {
		executeTemplate(w, "pplTracking", data)
	} else {
		if data.Message == "" {
			data.Message = "Zásilka nebyla nalezena"
		}
		executeTemplate(w, "pplPackageNotFound", data)
	}
}
