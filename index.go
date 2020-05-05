package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/blevesearch/bleve"
)

// Sound description d'un son de la soundbox
type Sound struct {
	Character string `json:"character"`
	Episode   string `json:"episode"`
	File      string `json:"file"`
	Title     string `json:"title"`
}

// SoundIndex Aide à la recherche de son
type SoundIndex struct {
	index        bleve.Index
	Sounds       []Sound
	soundsByFile map[string]Sound // soundsMap[fileName]
}

// Search recherche un son dans l'index
func (si *SoundIndex) Search(str string) ([]Sound, error) {
	var resp []Sound

	query := bleve.NewMatchQuery(str)
	req := bleve.NewSearchRequest(query)
	searchResults, err := si.index.Search(req)
	if err != nil {
		return nil, fmt.Errorf("erreur dans la requête > %w", err)
	}
	for _, hit := range searchResults.Hits {
		if sound, ok := si.soundsByFile[hit.ID]; ok {
			resp = append(resp, sound)
		}
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("pas de fichier trouvé")
	}

	return resp, nil
}

func loadAndIndexSounds(path string) (*SoundIndex, error) {

	var sounds []Sound
	soundMap := make(map[string]Sound)

	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base de sons > %w", err)
	}

	byteData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("lecture de la base de sons > %w", err)
	}

	if json.Unmarshal(byteData, &sounds) != nil {
		return nil, fmt.Errorf("lecture de la base des sons > %w", err)
	}

	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, fmt.Errorf("création de l'indexeur > %w", err)
	}

	for _, sound := range sounds {
		if index.Index(sound.File, sound) != nil {
			log.Printf("Error d'indexion de %v", sound)
		}
		soundMap[sound.File] = sound
	}

	return &SoundIndex{
		index,
		sounds,
		soundMap,
	}, nil
}
