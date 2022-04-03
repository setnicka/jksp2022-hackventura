package main

import (
	"net/http"
	"path"
	"sort"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/setnicka/jksp2022-hackventura/state"
)

func orgRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.BasicAuth("Jen pro orgy", map[string]string{orgLogin: orgPassword}))

	staticHandler := http.FileServer(NoListFileSystem{http.Dir(path.Join(staticDir, "/org"))})
	r.Mount("/static/", http.StripPrefix("/static/", staticHandler))

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Get("/", orgDashboardGet)
		r.Get("/teams", orgTeamsGet)
		r.Post("/teams", orgTeamsPost)
		r.Get("/settings", orgSettingsGet)
		r.Post("/settings", orgSettingsPost)
	})
	r.Get("/state-version", orgStateVersionGet)

	return r
}

func orgStateVersionGet(w http.ResponseWriter, r *http.Request) {
	version := server.state.Version
	w.Write([]byte(strconv.Itoa(version)))
}

////////////////////////////////////////////////////////////////////////////////

type orgTeamsData struct {
	GeneralData
	Teams []*state.Team
}

func orgTeamsGet(w http.ResponseWriter, r *http.Request) {
	server.state.RLock()
	defer server.state.RUnlock()
	data := orgTeamsData{
		GeneralData: getGeneralData("Týmy", w, r),
		Teams:       server.state.Teams,
	}
	sort.Slice(data.Teams, func(i, j int) bool {
		return data.Teams[i].Login < data.Teams[j].Login
	})
	executeTemplate(w, "orgTeams", data)
}

func orgTeamsPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/teams", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse teams form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	if r.PostFormValue("deleteTeam") != "" {
		login := r.PostFormValue("deleteTeam")
		if team := server.state.GetTeam(login); team != nil {
			if err := server.state.DeleteTeam(login); err != nil {
				setFlashMessage(w, r, messageError, "Chyba při mazání týmu: %v", err)
			} else {
				setFlashMessage(w, r, messageOk, "Tým smazán")
			}
		} else {
			setFlashMessage(w, r, messageWarn, "Tým '%s' nenalezen", login)
		}
	} else if r.PostFormValue("setPassword") != "" {
		login := r.PostFormValue("login")
		password := r.PostFormValue("setPassword")
		if team := server.state.GetTeam(login); team != nil {
			if err := server.state.TeamSetPassword(login, password); err != nil {
				setFlashMessage(w, r, messageError, "Chyba při změně hesla: %v", err)
			} else {
				setFlashMessage(w, r, messageOk, "Heslo nastaveno")
			}
		} else {
			setFlashMessage(w, r, messageWarn, "Tým '%s' nenalezen", login)
		}
	} else if r.PostFormValue("newTeamLogin") != "" {
		if err := server.state.AddTeam(r.PostFormValue("newTeamLogin"), r.PostFormValue("newTeamName")); err != nil {
			setFlashMessage(w, r, messageError, "Nelze přidat tým: %v", err)
		} else {
			setFlashMessage(w, r, messageOk, "Tým přidán")
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type teamServiceResult struct {
	Completed     bool
	CompletedTime string
	Tries         int
}

type teamResult struct {
	Name    string
	Results []teamServiceResult
}

type orgDashboardData struct {
	GeneralData
	Version int
	Targets []targetS
	Teams   []teamResult
}

func orgDashboardGet(w http.ResponseWriter, r *http.Request) {
	server.state.RLock()
	defer server.state.RUnlock()

	data := orgDashboardData{
		GeneralData: getGeneralData("Výsledky", w, r),
		Targets:     targets,
		Teams:       []teamResult{},
		Version:     server.state.Version,
	}

	for _, team := range server.state.Teams {
		res := teamResult{team.Name, []teamServiceResult{
			{team.CIA.Completed, team.CIA.CompletedTime.Format("15:04:05"), team.CIA.Tries},
			{team.NSA.Completed, team.NSA.CompletedTime.Format("15:04:05"), team.NSA.Tries},
			{team.KGB.Completed, team.KGB.CompletedTime.Format("15:04:05"), team.KGB.Tries},
			{team.FBI.Completed, team.FBI.CompletedTime.Format("15:04:05"), team.FBI.Tries},
			{team.PPL.Completed, team.PPL.CompletedTime.Format("15:04:05"), team.PPL.Tries},
			{team.BIS.Completed, team.BIS.CompletedTime.Format("15:04:05"), team.BIS.Tries},
			{team.MI5.Completed, team.MI5.CompletedTime.Format("15:04:05"), team.MI5.Tries},
		}}
		data.Teams = append(data.Teams, res)
	}

	executeTemplate(w, "orgDashboard", data)
}

////////////////////////////////////////////////////////////////////////////////

type orgSettingsData struct {
	GeneralData
	DolGuldor []state.DolGuldurPassword
}

func orgSettingsGet(w http.ResponseWriter, r *http.Request) {
	server.state.RLock()
	defer server.state.RUnlock()

	data := orgSettingsData{
		GeneralData: getGeneralData("Nastavení", w, r),
		DolGuldor:   server.state.GetDolGuldurPasswords(),
	}

	executeTemplate(w, "orgSettings", data)
}

func orgSettingsPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/settings", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse settings form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	if r.PostFormValue("dolGuldorPassword") != "" {
		password := r.PostFormValue("dolGuldorPassword")
		if err := server.state.SetDolGuldurPassword(password); err != nil {
			setFlashMessage(w, r, messageError, "Chyba při nastavování hesla: %v", err)
		} else {
			setFlashMessage(w, r, messageOk, "Nové heslo nastaveno")
		}
	}
}
