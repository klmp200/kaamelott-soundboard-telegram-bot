package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type settings struct {
	ListeningAddress string `json:"listening_address"`
	Domain           string `json:"domain"`
	TelegramKey      string `json:"telegram_key"`
}

const (
	soundDatabase           = "./sounds/sounds.json"
	soundFolderPrefix       = "./sounds"
	defaultlisteningAddress = "0.0.0.0:8080"
	defaultDomain           = "http://localhost:8080"
	defaultTelegramKey      = "nope"
)

var index *SoundIndex
var cfg *settings

func loadSettings(path string) (*settings, error) {
	loaded := settings{
		ListeningAddress: defaultlisteningAddress,
		Domain:           defaultDomain,
		TelegramKey:      defaultTelegramKey,
	}

	jsonFile, err := os.Open(path)
	if err != nil {
		log.Printf("non trouvé mais c'est pas grave on fait sans")
		return &loaded, nil
	}

	byteData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Je peux pas le lire ton fichier %s > %w", path, err)
	}

	if json.Unmarshal(byteData, &loaded) != nil {
		return nil, fmt.Errorf("Ton json %s il est tout pété > %w", path, err)
	}

	return &loaded, nil
}

func main() {

	cfg, err := loadSettings("settings.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	index, err := loadAndIndexSounds(soundDatabase)
	if err != nil {
		log.Fatal(err)
		return
	}

	b, err := tb.NewBot(tb.Settings{
		Token:  cfg.TelegramKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	fs := http.FileServer(http.Dir(soundFolderPrefix))
	http.Handle("/", fs)

	log.Printf("Écoute sur %s...", cfg.ListeningAddress)
	go http.ListenAndServe(cfg.ListeningAddress, nil)

	b.Handle("/cite", func(m *tb.Message) {
		sounds, err := index.Search(m.Text)
		if err != nil {
			b.Send(m.Chat, "Pas de fichier audio trouvé")
			return
		}
		path := fmt.Sprintf("%s/%s", soundFolderPrefix, sounds[0].File)
		_, err = b.Send(
			m.Chat,
			&tb.Audio{
				File:      tb.FromDisk(path),
				Caption:   sounds[0].Title,
				Title:     sounds[0].Episode,
				Performer: sounds[0].Character,
			},
		)
		if err != nil {
			log.Panicf("Impossible d'envoyer le fichier %s: %s", path, err)
			b.Send(m.Chat, "Erreur d'envoi du fichier")
		}
	})

	b.Handle(tb.OnQuery, func(q *tb.Query) {
		sounds, err := index.Search(q.Text)
		if err != nil {
			log.Println("Pas de fichier audio trouvé")
			return
		}
		results := make(tb.Results, len(sounds))
		for i, sound := range sounds {
			url := fmt.Sprintf("%s/%s", cfg.Domain, sound.File)
			log.Println(url)
			results[i] = &tb.AudioResult{
				URL:       url,
				Caption:   sound.Episode,
				Title:     sound.Title,
				Performer: sound.Character,
			}
			results[i].SetResultID(strconv.Itoa(i))
		}
		err = b.Answer(q, &tb.QueryResponse{
			Results:           results,
			SwitchPMParameter: "Ajoute moi",
			CacheTime:         60, // une minute
		})

		if err != nil {
			log.Println(err)
		}
	})

	b.Start()

}
