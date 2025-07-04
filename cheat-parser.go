package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Cheat struct {
	Note    string            `json:"note"`
	Data    []string          `json:"data"`
	Options map[string]string `json:"options"`
}

func isHex(s string) bool {
	var hexRegex = regexp.MustCompile(`^[0-9A-F ?]+$`)
	return hexRegex.MatchString(s)
}

func main() {
	fs := memfs.New()

	if _, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:   "https://github.com/project64/project64.git",
		Depth: 1,
	}); err != nil {
		log.Panic(err)
	}

	cheatDB := map[string]map[string]Cheat{}

	paths := []string{"Config/Cheats", "Config/Enhancements"}
	for _, basePath := range paths {
		items, err := fs.ReadDir(basePath)
		if err != nil {
			log.Panic(err)
		}
		for _, item := range items {
			currentGame := ""
			currentCheat := ""
			file, err := fs.Open(filepath.Join(basePath, item.Name()))
			if err != nil {
				log.Panic(err)
			}
			defer file.Close() //nolint:errcheck
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "[") {
					currentGame = strings.Trim(scanner.Text(), "[]")
					if cheatDB[currentGame] == nil {
						cheatDB[currentGame] = map[string]Cheat{}
					}
				} else if strings.HasPrefix(scanner.Text(), "Name=") {
					// do nothing
				} else if strings.HasPrefix(scanner.Text(), "$") {
					currentCheat = strings.TrimPrefix(scanner.Text(), "$")
					cheatDB[currentGame][currentCheat] = Cheat{}
				} else if strings.HasPrefix(scanner.Text(), "Note=") && currentCheat != "" {
					current := cheatDB[currentGame][currentCheat]
					current.Note = strings.TrimPrefix(scanner.Text(), "Note=")
					cheatDB[currentGame][currentCheat] = current
				} else if scanner.Text() == "" {
					// do nothing
				} else if strings.Contains(scanner.Text(), "?") && currentCheat != "" && cheatDB[currentGame][currentCheat].Options == nil {
					current := cheatDB[currentGame][currentCheat]
					current.Options = map[string]string{}
					current.Data = append(cheatDB[currentGame][currentCheat].Data, scanner.Text())
					cheatDB[currentGame][currentCheat] = current
				} else if currentCheat != "" && cheatDB[currentGame][currentCheat].Options != nil && len(strings.Split(scanner.Text(), " ")[0]) < 8 {
					cheatDB[currentGame][currentCheat].Options[strings.Join(strings.Split(scanner.Text(), " ")[1:], " ")] = strings.Split(scanner.Text(), " ")[0]
				} else if isHex(scanner.Text()) && currentCheat != "" {
					current := cheatDB[currentGame][currentCheat]
					current.Data = append(cheatDB[currentGame][currentCheat].Data, scanner.Text())
					cheatDB[currentGame][currentCheat] = current
				} else if currentCheat != "" && (strings.HasPrefix(scanner.Text(), "OnByDefault=1") || strings.HasPrefix(scanner.Text(), "PluginList=")) {
					// PJ64 specific, ignore
					delete(cheatDB[currentGame], currentCheat)
					log.Printf("Ignoring line in cheat file %s, %s: %s\n", filepath.Join(basePath, item.Name()), currentCheat, scanner.Text())
					currentCheat = ""
				} else if currentCheat != "" {
					delete(cheatDB[currentGame], currentCheat)
					log.Printf("Unknown line in cheat file %s, %s: %s\n", filepath.Join(basePath, item.Name()), currentCheat, scanner.Text())
					currentCheat = ""
				}
			}

			if err := scanner.Err(); err != nil {
				log.Panic(err)
			}
		}
	}

	for game, cheats := range cheatDB {
		for cheat_name, cheat := range cheats {
			if len(cheat.Data) == 0 {
				delete(cheatDB[game], cheat_name)
				log.Printf("Removing empty cheat %s for game %s\n", cheat_name, game)
			}
		}
	}

	b, err := json.Marshal(cheatDB)
	if err != nil {
		log.Panic(err)
	}

	f, err := os.OpenFile("cheats.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Panic(err)
	}

	w := bufio.NewWriter(f)
	if _, err := w.Write(b); err != nil {
		log.Panic(err)
	}
	w.Flush() //nolint:errcheck
}
