package main

//127.0.0.1:8080

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
	}
	mu sync.Mutex
)

func initGame() {
	mots := Hangman.ChargerMots("words.txt")
	gameState.Mot = Hangman.ChoisirMot(mots)
	gameState.MotRevele = Hangman.RevelerLettres(gameState.Mot, len(gameState.Mot)/2-1)
	gameState.EssaisRestants = 6
	gameState.LettresEssayees = make(map[rune]bool)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t := template.Must(template.ParseFiles("templates/" + tmpl + ".tmpl"))
	t.Execute(w, data)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Calculer le nombre d'erreurs
	erreurs := 6 - gameState.EssaisRestants

	data := struct {
		Word         string
		AttemptsLeft int
		TriedLetters string
		Errors       int // Ajouter le nombre d'erreurs
	}{
		Word:         Hangman.AfficherMotRevele(gameState.MotRevele),
		AttemptsLeft: gameState.EssaisRestants,
		TriedLetters: strings.Join(keys(gameState.LettresEssayees), ", "),
		Errors:       erreurs, // Passer le nombre d'erreurs
	}

	renderTemplate(w, "index", data)
}

func hangmanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	guess := r.FormValue("guess")
	if len(guess) != 1 || !strings.Contains("abcdefghijklmnopqrstuvwxyz", guess) {
		http.Error(w, "Invalide input", http.StatusBadRequest)
		return
	}

	char := rune(guess[0])
	if gameState.LettresEssayees[char] {
		http.Error(w, "Lettre déjà essayée", http.StatusBadRequest)
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
		message := "Félicitations, vous avez gagné !"
		resultClass := "success"
		if gameState.EssaisRestants == 0 {
			message = "Vous avez perdu.\nLe mot était : " + gameState.Mot
			resultClass = "failure"
		}
		initGame()
		renderTemplate(w, "result", struct {
			Message     string
			ResultClass string
		}{
			Message:     message,
			ResultClass: resultClass,
		})
		initGame()
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func keys(m map[rune]bool) []string {
	var result []string
	for k := range m {
		result = append(result, string(k))
	}
	return result
}

func main() {
	initGame()
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hangman", hangmanHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("Le serveur est en cours d'exécution sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
