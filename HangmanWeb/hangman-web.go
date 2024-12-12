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
	mu sync.Mutex
)

func initGame() {
	mots := Hangman.ChargerMots("words.txt")
	etapes := Hangman.ChargerPendu("hangman.txt")
	gameState.Mot = Hangman.ChoisirMot(mots)
	gameState.MotRevele = Hangman.RevelerLettres(gameState.Mot, len(gameState.Mot)/2-1)
	gameState.EssaisRestants = 6
	gameState.LettresEssayees = make(map[rune]bool)
	gameState.EtapesPendu = etapes
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t := template.Must(template.ParseFiles("templates/" + tmpl + ".tmpl"))
	fmt.Println(t)
	t.Execute(w, data)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Cette route affiche la page de démarrage
	fmt.Println("hello")
	renderTemplate(w, "index", nil)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("bjr")
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
	}{
		Word:         Hangman.AfficherMotRevele(gameState.MotRevele),
		AttemptsLeft: gameState.EssaisRestants,
		TriedLetters: strings.Join(keys(gameState.LettresEssayees), ", "),
		Errors:       erreurs,
		PenduImage:   gameState.EtapesPendu[6-gameState.EssaisRestants],
	}

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
		http.Error(w, "Entrée invalide", http.StatusBadRequest)
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

		renderTemplate(w, "result", struct {
			Message     string
			ResultClass string
		}{
			Message:     message,
			ResultClass: resultClass,
		})

		// Initialisation du jeu après la fin de la partie
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

	// Gestion des routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/hangman", hangmanHandler)

	// Gérer les ressources statiques
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Lancer le serveur
	fmt.Println("Le serveur est en cours d'exécution sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
