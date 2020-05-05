package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/blevesearch/bleve"
)

// Sound description d'un son
type Sound struct {
	Character string `json:"character"`
	Episode   string `json:"episode"`
	File      string `json:"file"`
	Title     string `json:"title"`
}

func main() {

	var sounds []Sound

	jsonFile, err := os.Open("sounds.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	byteData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	if json.Unmarshal(byteData, &sounds) != nil {
		log.Fatal(err)
		return
	}

	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, sound := range sounds {
		if index.Index(sound.File, sound) != nil {
			log.Printf("Error d'indexion de %v", sound)
		}
	}

	query := bleve.NewMatchQuery("cul")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, file := range searchResults.Hits {
		fmt.Println(file.ID)
	}

}
