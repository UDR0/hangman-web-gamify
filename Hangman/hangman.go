package Hangman

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func ChargerMots(fichier string) []string {
	file, err := os.Open(fichier)
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier :", err)
		os.Exit(1)
	}
	defer file.Close()

	var mots []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		mots = append(mots, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Erreur lors de la lecture du fichier :", err)
		os.Exit(1)
	}

	return mots
}
func ChargerPendu(fichier string) []string {
	file, err := os.Open(fichier)
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier :", err)
		os.Exit(1)
	}
	defer file.Close()

	var etapes []string
	var etape string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ligne := scanner.Text()
		if ligne == "" {
			etapes = append(etapes, etape)
			etape = ""
		} else {
			etape += ligne + "\n"
		}
	}
	if etape != "" {
		etapes = append(etapes, etape)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Erreur lors de la lecture du fichier :", err)
		os.Exit(1)
	}

	return etapes
}
func ChoisirMot(mots []string) string {
	rand.Seed(time.Now().UnixNano())
	return mots[rand.Intn(len(mots))]
}
func RevelerLettres(mot string, n int) []rune {
	motRevele := make([]rune, len(mot))
	for i := range motRevele {
		motRevele[i] = '_'
	}

	indicesReveles := rand.Perm(len(mot))[:n]
	for _, indice := range indicesReveles {
		motRevele[indice] = rune(mot[indice])
	}
	return motRevele
}
func AfficherMotRevele(motRevele []rune) string {
	return strings.Join(strings.Split(string(motRevele), ""), " ")
}
func AfficherPendu(etapes []string, nbEssais int) {
	fmt.Println(etapes[6-nbEssais])
}
func JouerPendu(mot string, etapes []string) {
	nbEssais := 6
	motRevele := RevelerLettres(mot, len(mot)/2-1)
	lettresEssayees := make(map[rune]bool)

	for nbEssais > 0 {
		AfficherPendu(etapes, nbEssais)
		fmt.Println("Mot à deviner :", AfficherMotRevele(motRevele))
		fmt.Println("Essais restants :", nbEssais)
		fmt.Print("Entrez une lettre : ")
		var lettre string
		fmt.Scanln(&lettre)
		lettre = strings.ToLower(lettre)
		if len(lettre) != 1 || !strings.Contains("abcdefghijklmnopqrstuvwxyz", lettre) {
			fmt.Println("Veuillez entrer une seule lettre valide.")
			continue
		}
		char := rune(lettre[0])
		if lettresEssayees[char] {
			fmt.Println("Vous avez déjà essayé cette lettre.")
			continue
		}
		lettresEssayees[char] = true
		if strings.Contains(mot, lettre) {
			fmt.Println("Bravo ! La lettre", lettre, "est dans le mot.")
			for i, lettreMot := range mot {
				if lettreMot == char {
					motRevele[i] = char
				}
			}
		} else {
			fmt.Println("Dommage, la lettre", lettre, "n'est pas dans le mot.")
			nbEssais--
		}
		if string(motRevele) == mot {
			fmt.Println("Félicitations ! Vous avez trouvé le mot :", mot)
			return
		}
	}
	AfficherPendu(etapes, 0)
	fmt.Println("Désolé, vous avez perdu. Le mot était :", mot)
}
func main() {
	mots := ChargerMots("words.txt")
	etapes := ChargerPendu("hangman.txt")
	mot := ChoisirMot(mots)
	JouerPendu(mot, etapes)
}
