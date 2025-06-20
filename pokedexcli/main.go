package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jfallin97/pokedexcli/internal"
)

// RNG
var rng *rand.Rand

// Pokedex storage map - for when they catch a pokemon
var PokeBox = make(map[string]ActualPokemon)

// Global cache var
var cache *internal.Cache

// Offset is a part of the api url we call in AreaLocation
var currentOffset = 0

type PokeAreaResponse struct {
	Count    int              `json:"count"`
	Next     *string          `json:"next"`
	Previous *string          `json:"previous"`
	Results  []ActualPokeArea `json:"results"`
}

type EncounterResponse struct {
	Encounters []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonEncounter struct {
	Pokes Pokemon `json:"pokemon"`
}

type Pokemon struct {
	Name string `json:"name"`
}

type ActualPokeArea struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type cliCommand struct {
	name        string
	description string
	callback    func([]string) error
}

type ActualPokemon struct {
	Name           string      `json:"name"`
	BaseExperience int         `json:"base_experience"`
	Height         int         `json:"height"`
	Weight         int         `json:"weight"`
	Stats          []PokeStats `json:"stats"`
	Types          []PokeTypes `json:"types"`
}

type PokeStats struct {
	BaseStat int       `json:"base_stat"`
	Stat     DeepStats `json:"stat"`
}

type DeepStats struct {
	Name string `json:"name"`
}

type TypeInfo struct {
	Name string `json:"name"`
}

type PokeTypes struct {
	Types TypeInfo `json:"type"`
}

var commandList map[string]cliCommand

func init() {
	// RNG
	source := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(source)

	// Commands
	commandList = map[string]cliCommand{
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
			description: "Displays the names of 20 location areas.",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore what pokemon are in a given area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to Catch a Pokemon!",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "View information and stats about the pokemon. (you must have caught it)",
			callback:    commandInspect,
		},
	}
}

func main() {

	// init cache via our func from pokecache
	cache = internal.NewCache(5 * time.Second)
	// Bust out the scanner
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		command := scanner.Text()
		cleaned := cleanInput(command)
		userCommand, exists := commandList[cleaned[0]]
		if exists {
			userCommand.callback(cleaned[1:])
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func cleanInput(text string) []string {
	var input []string
	var final []string
	for _, s := range strings.Fields(text) {
		input = append(input, s)
	}
	for _, i := range input {
		final = append(final, strings.ToLower(i))
	}
	return final
}

func commandExit(args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(args []string) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	for _, command := range commandList {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandMap(args []string) error {
	MapUrl := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/?offset=%d&limit=20", currentOffset)

	// Check cache first
	cachedData, found := cache.Get(MapUrl)
	var B_bytes []byte

	if found {
		B_bytes = cachedData
	} else {
		res, err := http.Get(MapUrl)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var err2 error
		B_bytes, err2 = io.ReadAll(res.Body)
		if err2 != nil {
			log.Fatal(err2)
		}

		// Store in cache
		cache.Add(MapUrl, B_bytes)
	}

	var area PokeAreaResponse
	errs := json.Unmarshal(B_bytes, &area)
	if errs != nil {
		log.Fatal(errs)
	}

	for _, i := range area.Results {
		fmt.Println(i.Name)
	}

	currentOffset += 20
	return nil
}

func commandMapb(args []string) error {
	MapUrl := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/?offset=%d&limit=20", currentOffset)

	// Check cache first
	cachedData, found := cache.Get(MapUrl)
	var B_bytes []byte

	if found {
		B_bytes = cachedData
	} else {
		res, err := http.Get(MapUrl)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var err2 error
		B_bytes, err2 = io.ReadAll(res.Body)
		if err2 != nil {
			log.Fatal(err2)
		}

		// Store in cache
		cache.Add(MapUrl, B_bytes)
	}

	var area PokeAreaResponse
	errs := json.Unmarshal(B_bytes, &area)
	if errs != nil {
		log.Fatal(errs)
	}

	for _, i := range area.Results {
		fmt.Println(i.Name)
	}

	currentOffset -= 20
	return nil
}

func commandExplore(args []string) error {

	// check for invalid input
	if len(args) == 0 {
		fmt.Println("Proper area input required")
		return nil
	}

	// take in area name
	areaName := args[0]
	fmt.Printf("Exploring %s...\n", areaName)
	fmt.Printf("Found Pokemon:\n")

	MapUrl := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%v", areaName)

	// Check cache first
	cachedData, found := cache.Get(MapUrl)
	var B_bytes []byte

	if found {
		B_bytes = cachedData
	} else {
		res, err := http.Get(MapUrl)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var err2 error
		B_bytes, err2 = io.ReadAll(res.Body)
		if err2 != nil {
			log.Fatal(err2)
		}

		// Store in cache
		cache.Add(MapUrl, B_bytes)
	}

	var PokeList EncounterResponse
	errs := json.Unmarshal(B_bytes, &PokeList)
	if errs != nil {
		log.Fatal(errs)
	}

	for _, i := range PokeList.Encounters {
		fmt.Println(i.Pokes.Name)
	}

	return nil
}

func commandCatch(args []string) error {
	// check for invalid input
	if len(args) == 0 {
		fmt.Println("Proper area input required")
		return nil
	}

	RequestedPoke := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", RequestedPoke)
	MapUrl := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", args[0])

	// Check cache first
	cachedData, found := cache.Get(MapUrl)
	var B_bytes []byte

	if found {
		B_bytes = cachedData
	} else {
		res, err := http.Get(MapUrl)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var err2 error
		B_bytes, err2 = io.ReadAll(res.Body)
		if err2 != nil {
			log.Fatal(err2)
		}

		// Store in cache
		cache.Add(MapUrl, B_bytes)
	}

	var monster ActualPokemon

	errs := json.Unmarshal(B_bytes, &monster)
	if errs != nil {
		log.Fatal(errs)
	}

	var PokeName = monster.Name
	var BaseExperience = monster.BaseExperience
	var CatchChance = 100 - BaseExperience

	if caughtOrNot(CatchChance) {
		fmt.Printf("%s was caught!\n", PokeName)
		PokeBox[PokeName] = monster
	} else {
		fmt.Printf("%s got away!\n", PokeName)
	}

	return nil
}

// Returns true if successful catch. We give it a chance to catch.
func caughtOrNot(chance int) bool {
	return rng.Intn(100) < chance
}

// func for checking a pokemons stats and information
func commandInspect(args []string) error {
	if len(args) == 0 {
		fmt.Println("Please specify a Pokemon name to inspect")
		return nil
	}
	var pokemonName = args[0]

	pokemon, exists := PokeBox[pokemonName]
	if !exists {
		fmt.Printf("You haven't caught %s yet!\n", pokemonName)
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf(" - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, typeInfo := range pokemon.Types {
		fmt.Printf(" - %s\n", typeInfo.Types.Name)
	}

	return nil
}
