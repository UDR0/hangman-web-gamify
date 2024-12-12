package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"example.com/Hangman"
)

var (
	gameState struct {
		Mot             string
		MotRevele       []rune
		EssaisRestants  int
		LettresEssayees map[rune]bool
		EtapesPendu     []string
	}
	mu           sync.Mutex
	errorMessage string
)

func initGame() {
	mots := Hangman.ChargerMots("words.txt")
	etapes := Hangman.ChargerPendu("hangman.txt")
	gameState.Mot = Hangman.ChoisirMot(mots)
	gameState.MotRevele = Hangman.RevelerLettres(gameState.Mot, len(gameState.Mot)/2-1)
	gameState.EssaisRestants = 6
	gameState.LettresEssayees = make(map[rune]bool)
	gameState.EtapesPendu = etapes
	errorMessage = ""
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t := template.Must(template.ParseFiles("templates/" + tmpl + ".tmpl"))
	t.Execute(w, data)
}

func keys(m map[rune]bool) []string {
	var result []string
	for k := range m {
		result = append(result, string(k))
	}
	return result
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Calculer le nombre d'erreurs
	erreurs := 6 - gameState.EssaisRestants

	data := struct {
		Word         string
		AttemptsLeft int
		TriedLetters string
		Errors       int
		PenduImage   string
		ErrorMessage string
	}{
		Word:         Hangman.AfficherMotRevele(gameState.MotRevele),
		AttemptsLeft: gameState.EssaisRestants,
		TriedLetters: strings.Join(keys(gameState.LettresEssayees), ", "),
		Errors:       erreurs,
		PenduImage:   gameState.EtapesPendu[6-gameState.EssaisRestants],
		ErrorMessage: errorMessage,
	}

	errorMessage = "" // Réinitialiser le message d'erreur après affichage
	renderTemplate(w, "start", data)
}

func hangmanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/hangman", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	guess := r.FormValue("guess")
	if len(guess) != 1 || !strings.Contains("abcdefghijklmnopqrstuvwxyz", guess) {
		errorMessage = "Entrée invalide. Veuillez entrer une seule lettre minuscule."
		http.Redirect(w, r, "/start", http.StatusSeeOther)
		return
	}

	char := rune(guess[0])
	if gameState.LettresEssayees[char] {
		errorMessage = "Lettre déjà essayée."
		http.Redirect(w, r, "/start", http.StatusSeeOther)
		return
	}

	gameState.LettresEssayees[char] = true
	if strings.Contains(gameState.Mot, guess) {
		for i, lettre := range gameState.Mot {
			if lettre == char {
				gameState.MotRevele[i] = char
			}
		}
	} else {
		gameState.EssaisRestants--
	}

	if gameState.EssaisRestants == 0 || string(gameState.MotRevele) == gameState.Mot {
		message := "Félicitations, vous avez gagné ! Le mot était : " + gameState.Mot
		resultClass := "success"
		if gameState.EssaisRestants == 0 {
			message = "Vous avez perdu. Le mot était : " + gameState.Mot
			resultClass = "failure"
		}

		renderTemplate(w, "result", struct {
			Message     string
			ResultClass string
		}{
			Message:     message,
			ResultClass: resultClass,
		})

		// Réinitialiser le jeu après la fin de la partie
		initGame()
		return
	}

	http.Redirect(w, r, "/start", http.StatusSeeOther)
}

func main() {
	initGame()

	// Routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/hangman", hangmanHandler)

	// Gérer les ressources statiques
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("Serveur démarré sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
