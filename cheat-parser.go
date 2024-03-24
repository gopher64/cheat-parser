package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Cheat struct {
	Note       string            `json:"note"`
	Data       []string          `json:"data"`
	Options    map[string]string `json:"options"`
	HasOptions bool              `json:"hasOptions"`
}

func main() {
	fs := memfs.New()

	if _, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:   "https://github.com/project64/project64.git",
		Depth: 1,
	}); err != nil {
		log.Panic(err)
	}

	basePath := "Config/Cheats"
	items, err := fs.ReadDir(basePath)
	if err != nil {
		log.Panic(err)
	}

	cheatDB := map[string]map[string]Cheat{}
	currentGame := ""
	currentCheat := ""
	for _, item := range items {
		file, err := fs.Open(filepath.Join(basePath, item.Name()))
		if err != nil {
			log.Panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			game := cheatDB[currentGame]
			cheat := game[currentCheat]
			if strings.HasPrefix(scanner.Text(), "[") {
				currentGame = strings.Trim(scanner.Text(), "[]")
				cheatDB[currentGame] = map[string]Cheat{}
			} else if strings.HasPrefix(scanner.Text(), "Name=") {
				// do nothing
			} else if strings.HasPrefix(scanner.Text(), "$") {
				currentCheat = strings.TrimPrefix(scanner.Text(), "$")
				game[currentCheat] = Cheat{}
				cheatDB[currentGame] = game
			} else if strings.HasPrefix(scanner.Text(), "Note=") {
				cheat.Note = strings.TrimPrefix(scanner.Text(), "Note=")
				game[currentCheat] = cheat
			} else if scanner.Text() == "" {
				// do nothing
			} else if strings.Contains(scanner.Text(), "?") && !cheat.HasOptions {
				cheat.HasOptions = true
				cheat.Options = map[string]string{}
				cheat.Data = append(cheat.Data, scanner.Text())
				game[currentCheat] = cheat
			} else if cheat.HasOptions && len(strings.Split(scanner.Text(), " ")[0]) < 8 {
				cheat.Options[strings.Join(strings.Split(scanner.Text(), " ")[1:], " ")] = strings.Split(scanner.Text(), " ")[0]
				game[currentCheat] = cheat
			} else {
				cheat.Data = append(cheat.Data, scanner.Text())
				game[currentCheat] = cheat
			}
		}

		if err := scanner.Err(); err != nil {
			log.Panic(err)
		}
	}

	b, err := json.Marshal(cheatDB)
	if err != nil {
		log.Panic(err)
	}

	f, err := os.OpenFile("cheats.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644) //nolint:gomnd
	if err != nil {
		log.Panic(err)
	}

	w := bufio.NewWriter(f)
	if _, err := w.Write(b); err != nil {
		log.Panic(err)
	}
	w.Flush()
}
