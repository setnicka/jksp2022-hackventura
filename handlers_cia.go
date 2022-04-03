package main

import (
	"net/http"
	"regexp"
	"time"
	"unicode"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	ciaLogin     = "agentfred"
	ciaTimeLimit = time.Second * 1 // mezi pokusy o heslo
)

func ciaRouter() *chi.Mux {
	r := newGameRouter("cia")
	r.Get("/", auth(ciaIndexGet))
	r.Post("/", auth(ciaIndexPost))
	r.Get("/internal", auth(ciaInternalGet))
	return r
}

func ciaIndexGet(w http.ResponseWriter, r *http.Request) {
	data := getGeneralData("CIA", w, r)
	defer func() { executeTemplate(w, "ciaIndex", data) }()

	team := server.state.GetTeam(getUser(r))
	if team != nil && team.CIA.Completed {
		http.Redirect(w, r, "/internal", http.StatusSeeOther)
		return
	}
}

func ciaIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if time.Since(team.CIA.LastTry) < ciaTimeLimit {
		setFlashMessage(w, r, messageError, "Minimal time between two login attempts is 1 seconds")
		return
	}

	team.CIA.Tries++
	team.CIA.LastTry = time.Now()
	defer server.state.Save()

	// e.F.g.H.i.J

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[CIA - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if login != ciaLogin {
		setFlashMessage(w, r, messageError, "Incorrect login name, only agents can log in")
		return
	} else if len(password) < 11 {
		setFlashMessage(w, r, messageError, "Incorrect password, password is too short")
		return
	} else if len(password) > 11 {
		setFlashMessage(w, r, messageError, "Incorrect password, password is too long")
		return
	}

	bpassword := []byte(password)

	reLetter := regexp.MustCompile("[a-z]")
	reBigLetter := regexp.MustCompile("[A-Z]")
	reNumbers := regexp.MustCompile("[0-9]")
	reDoubleLetters := regexp.MustCompile("[a-zA-Z][a-zA-Z]")

	smallLetters := len(reLetter.FindAll(bpassword, -1))
	bigLetters := len(reBigLetter.FindAll(bpassword, -1))
	doubleLetters := len(reDoubleLetters.FindAll(bpassword, -1))
	numbers := len(reNumbers.FindAll(bpassword, -1))
	letters := smallLetters + bigLetters
	other := len(password) - letters - numbers

	if letters < 3 {
		setFlashMessage(w, r, messageError, "Incorrect password, there isn't enough letters")
		return
	}

	if other < 3 {
		setFlashMessage(w, r, messageError, "Incorrect password, there are too many letters and numbers")
		return
	}
	if doubleLetters > 0 {
		setFlashMessage(w, r, messageError, "Incorrect password, two letters side by side are forbidden for safety reasons")
		return
	}
	if smallLetters != bigLetters {
		setFlashMessage(w, r, messageError, "Incorrect password, the number of lowercase letters differs from the number of uppercase letters")
		return
	}
	if other < numbers {
		setFlashMessage(w, r, messageError, "Incorrect password, there is too many numbers")
		return
	}

	req := r
	// pismena nesmi byt stejna
	wasLetter := map[rune]bool{}
	for _, r := range password {
		if unicode.IsLetter(r) {
			if _, found := wasLetter[r]; found {
				setFlashMessage(w, req, messageError, "Incorrect password, there cannot be two same letters")
				return
			}
			wasLetter[r] = true
		}
	}

	// pismena nejsou v abecednim poradi
	lastLetter := 'a'
	for _, r := range password {
		if unicode.IsLetter(r) {
			r = unicode.ToLower(r)
			if r < lastLetter {
				setFlashMessage(w, req, messageError, "Incorrect password, letters must be in alphabetical order")
				return
			}
			lastLetter = r
		}
	}

	// na tretim miste neni f
	if password[2] != 'f' && password[2] != 'F' {
		setFlashMessage(w, r, messageError, "Incorrect password, there isn't the first letter of your first name on the third position in the password")
		return
	}

	reAroundNumber := regexp.MustCompile("[a-zA-Z][0-9][a-zA-Z]")
	reAroundNumberSpecial := regexp.MustCompile("[^0-9a-zA-Z][0-9][^0-9a-zA-Z]")

	// cislo nesmi mit z obou stran specialni znak
	if reAroundNumberSpecial.Match(bpassword) {
		setFlashMessage(w, r, messageError, "Incorrect password, number cannot be surrounded by special characters from both sides")
		return
	}

	// cislo nesmi mit z obou stran pismeno
	if reAroundNumber.Match(bpassword) {
		setFlashMessage(w, r, messageError, "Incorrect password, number cannot be surrounded by letters from both sides")
		return
	}

	log.Infof("[CIA - %s] Completed", team.Login)
	// Everything completed
	team.CIA.Completed = true
	team.CIA.CompletedTime = time.Now()
	http.Redirect(w, r, "/internal", http.StatusSeeOther)
}

func ciaInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.CIA.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("CIA", w, r)
	executeTemplate(w, "ciaInternal", data)
}
