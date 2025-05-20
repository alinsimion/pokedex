package main

import (
	"bufio"

	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alinsimion/pokedex/internal/pokecache"
)

type Config struct {
	Next     *string
	Previous *string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
}

type Pokemon struct {
	Xp     int              `json:"base_experience"`
	Name   string           `json:"name"`
	Height int              `json:"height"`
	Weight int              `json:"weight"`
	Stats  []map[string]any `json:"stats"`
	Types  []map[string]any `json:"type"`
}

func (p Pokemon) String() string {

	tempString := fmt.Sprintf(`Name: %s
Height: %d
Weight: %d
Stats: %v
Types: %v
`, p.Name, p.Height, p.Weight, p.Stats, p.Types)

	return tempString
}

var (
	coms    map[string]cliCommand
	cache   = pokecache.NewCache(5 * time.Second)
	pokedex = make(map[string]Pokemon)
)

func main() {

	coms = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays 20 names",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays previous 20 names",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "Information about an area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catching a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspects a pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Lists all your pokemons",
			callback:    commandPokedex,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	config := Config{}

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()

		clnInput := cleanInput(input)

		commandName := clnInput[0]
		var args []string

		if len(clnInput) > 1 {
			args = clnInput[1:]
		}

		if command, ok := coms[commandName]; ok {
			err := command.callback(&config, args)

			if err != nil {
				fmt.Println(err.Error())
			}

		} else {
			fmt.Println("Unknown command")
		}
	}
}

func commandPokedex(config *Config, args []string) error {

	if len(pokedex) < 1 {
		fmt.Println("No pokemon")
	}

	for name := range pokedex {
		fmt.Println(name)
	}

	return nil
}

func commandInspect(config *Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("need at least an argument, the location area")
	}

	pokemon, ok := pokedex[args[0]]

	if !ok {
		fmt.Printf("%s is not in your pokedex\n", args[0])
		return nil
	}

	fmt.Printf("%v", pokemon)

	return nil
}

func commandCatch(config *Config, args []string) error {
	var url string
	var pokemon Pokemon

	if len(args) < 1 {
		return fmt.Errorf("need at least an argument, the location area")
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", args[0])

	url = fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", args[0])

	result, err := http.Get(url)

	if err != nil {
		slog.Error("Error while getting", "error", err.Error())
		return err
	}

	jsonBody, err := io.ReadAll(result.Body)

	if err != nil {
		slog.Error("Error while reading json body", "error", err.Error())
		return err
	}

	err = json.Unmarshal(jsonBody, &pokemon)

	if err != nil {
		slog.Error("Error while unmarshaling body", "error", err.Error())
		return err
	}

	chance := rand.IntN(200)

	if chance > pokemon.Xp {
		fmt.Printf("%s was caught!\n", pokemon.Name)
		pokedex[pokemon.Name] = pokemon
	} else {
		fmt.Printf("%s escaped!\n", pokemon.Name)
	}

	return nil
}

func commandExplore(config *Config, args []string) error {
	var url string

	if len(args) < 1 {
		return fmt.Errorf("need at least an argument, the location area")
	}

	var areaPokemonJson struct {
		Encounters []struct {
			Pokemon struct {
				Name string `json:"name"`
			} `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}

	url = "https://pokeapi.co/api/v2/location-area/" + args[0]

	result, err := http.Get(url)

	if err != nil {
		slog.Error("Error while getting", "error", err.Error())
		return err
	}

	jsonBody, err := io.ReadAll(result.Body)

	if err != nil {
		slog.Error("Error while reading json body", "error", err.Error())
		return err
	}

	err = json.Unmarshal(jsonBody, &areaPokemonJson)

	if err != nil {
		slog.Error("Error while unmarshaling body", "error", err.Error())
		return err
	}

	for _, enc := range areaPokemonJson.Encounters {
		fmt.Println(enc.Pokemon.Name)
	}

	return nil
}

func mapRequest(url string, config *Config) error {
	var areasJson struct {
		Count    int    `json:"count"`
		Next     string `json:"next"`
		Previous string `json:"previous"`
		Results  []struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"results"`
	}

	smth, ok := cache.Get(url)

	if !ok {
		result, err := http.Get(url)

		if err != nil {
			slog.Error("Error while getting", "error", err.Error())
			return err
		}

		jsonBody, err := io.ReadAll(result.Body)

		if err != nil {
			slog.Error("Error while reading json body", "error", err.Error())
			return err
		}

		cache.Add(url, jsonBody)

		smth = jsonBody
	}

	err := json.Unmarshal(smth, &areasJson)

	if err != nil {
		slog.Error("Error while unmarshaling body", "error", err.Error())
		return err
	}

	config.Next = &areasJson.Next
	config.Previous = &areasJson.Previous

	for _, elem := range areasJson.Results {
		fmt.Println(elem.Name)
	}

	return nil
}

func commandMap(config *Config, args []string) error {
	var url string

	if config.Next == nil {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = *config.Next
	}

	return mapRequest(url, config)
}

func commandMapB(config *Config, args []string) error {
	var url string

	if config.Previous == nil {
		fmt.Println("you're on the first page")
		return nil
	} else {

		if *config.Previous == "" {
			fmt.Println("you're on the first page")
			return nil
		}
		url = *config.Previous
	}

	return mapRequest(url, config)
}

func commandHelp(config *Config, args []string) error {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\n")
	for _, elem := range coms {
		fmt.Printf("%s: %s\n", elem.name, elem.description)
	}
	return nil
}

func commandExit(config *Config, args []string) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func cleanInput(text string) []string {

	var splits []string
	parts := strings.Split(text, " ")

	for _, elem := range parts {

		if elem != "" {
			splits = append(splits, strings.ToLower(elem))
		}

	}

	return splits
}
