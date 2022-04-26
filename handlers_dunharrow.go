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
	secondsBetweenAttempts = 11
	dunharrowTimeLimit = time.Second * secondsBetweenAttempts // mezi pokusy o heslo
)

func dunharrowRouter() *chi.Mux {
	r := newGameRouter("dunharrow")
	r.Get("/", auth(dunharrowIndexGet))
	r.Post("/", auth(dunharrowIndexPost))
	r.Get("/internal", auth(dunharrowInternalGet))
	return r
}

func dunharrowIndexGet(w http.ResponseWriter, r *http.Request) {
	data := getGeneralData("Dunharrow", w, r)
	defer func() { executeTemplate(w, "dunharrowIndex", data) }()

	team := server.state.GetTeam(getUser(r))
	if team != nil && team.Dunharrow.Completed {
		http.Redirect(w, r, "/internal", http.StatusSeeOther)
		return
	}
}

func dunharrowIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)

	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse spell form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	if time.Since(team.Dunharrow.LastTry) < dunharrowTimeLimit {
		setFlashMessage(w, r, messageError, "Po posledním kouzle jsi stále příliš slabý. Je třeba mezi kouzly čekat alespoň 11 sekund.")
		return
	}

	team.Dunharrow.Tries++
	team.Dunharrow.LastTry = time.Now()
	defer server.state.Save()

	// e.F.g.H.i.J

	password := r.PostFormValue("password")
	log.Infof("[Dunharrow - %s] Trying password '%s'", team.Login, password)

	if len(password) < 11 {
		setFlashMessage(w, r, messageError, "Tvé kouzlo bylo příliš slabé, aby prolomilo kletbu. Možná zafunguje nějaké s více znaky.")
		return
	} else if len(password) > 11 {
		setFlashMessage(w, r, messageError, "Seslané kouzlo by zajisté bylo dost silné, jenže je tak komplikované, že se ti jej nepodařilo vyřknout správně. Zkus něco kratšího.")
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
		setFlashMessage(w, r, messageError, "S duchy se nic nestalo. Nejspíš je v tvém kouzle málo písmenek.")
		return
	}

	if other < 3 {
		setFlashMessage(w, r, messageError, "Tvé kouzlo je příliš matoucí. Nejpíše obsahuje moc písmenek a číslic.")
		return
	}
	if doubleLetters > 0 {
		setFlashMessage(w, r, messageError, "Dvě písmena vedle sebe jsou příliš složitá na vyslovení, a tak se při kouzlení zakoktáš.")
		return
	}
	if smallLetters != bigLetters {
		setFlashMessage(w, r, messageError, "Kouzlo je nestabilní. K stabilitě by pomohlo, aby bylo stejně velkých jako malých písmen.")
		return
	}
	if other < numbers {
		setFlashMessage(w, r, messageError, "Duchové ti neporozuměli. Vyřkl jsi moc číslic.")
		return
	}

	req := r
	// pismena nesmi byt stejna
	wasLetter := map[rune]bool{}
	for _, r := range password {
		if unicode.IsLetter(r) {
			if _, found := wasLetter[r]; found {
				setFlashMessage(w, req, messageError, "Vyslovovat v kouzle dvě stejná písmena ti přijde moc nekreativní. Takže se o kouzlo ani nepokoušíš.")
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
				setFlashMessage(w, req, messageError, "Zapomněl jsi, v jakém pořadí jdou písmenka v kouzle za sebou. Raději je seřad abecedně, aby se to znovu nestalo.")
				return
			}
			lastLetter = r
		}
	}

	reGandalf := regexp.MustCompile("[gandlfGANDLF]")
	reSauron := regexp.MustCompile("[sauronSAURON]")
	reAroundNumber := regexp.MustCompile("[a-zA-Z][0-9][a-zA-Z]")
	reAroundNumberSpecial := regexp.MustCompile("[^0-9a-zA-Z][0-9][^0-9a-zA-Z]")

	gandalfLetters := len(reGandalf.FindAll(bpassword, -1))
	sauronLetters := len(reSauron.FindAll(bpassword, -1))

	// heslo má málo znaků z Gandalfa.
	if gandalfLetters < 2 {
		setFlashMessage(w, r, messageError, "Kouzlo nezvládneš sám. Budeš potřebovat Gandalfovu pomoc. Na tu máš ale v kouzle příliš málo znaků z jeho jména.")
		return
	}
	
	// heslo má moc znaků ze Saurona
	if sauronLetters > 0 {
		setFlashMessage(w, r, messageError, "I když už je Sauron mrtvý, jeho jméno stále vyvolává strach. Jelikož tvé kouzlo obsahuje příliš mnoho znaků z jeho jména, měl jsi při kouzlení děs v očích a kouzlo se nepovedlo.")
		return
	}

	// cislo nesmi mit z obou stran specialni znak
	if reAroundNumberSpecial.Match(bpassword) {
		setFlashMessage(w, r, messageError, "Jakmile je číslice obklopena speciálními znaky z obou stran, je to příliš kostrbaté na vyslovení. Zkus něco jiného.")
		return
	}

	// cislo nesmi mit z obou stran pismeno
	if reAroundNumber.Match(bpassword) {
		setFlashMessage(w, r, messageError, "Číslice obklopená písmenky z obou stran činí kouzlo neúčinné. Zkus tuto část kouzla změnit.")
		return
	}

	log.Infof("[Dunharrow - %s] Completed", team.Login)
	// Everything completed
	team.Dunharrow.Completed = true
	team.Dunharrow.CompletedTime = time.Now()
	http.Redirect(w, r, "/internal", http.StatusSeeOther)
}

func dunharrowInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Dunharrow.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Dunharrow", w, r)
	executeTemplate(w, "dunharrowInternal", data)
}
