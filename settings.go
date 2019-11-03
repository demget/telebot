package telebot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"text/template"
)

// NewSettings does try to load Settings from your json config file.
// 	- path is config path
// 	- dir is templates dir
func NewSettings(path, dir string) (Settings, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}

	var pref Settings
	if err := json.Unmarshal(data, &pref); err != nil {
		return Settings{}, err
	}
	if err := json.Unmarshal(data, &pref.Content); err != nil {
		return Settings{}, err
	}

	tmpl, err := template.New("data").ParseGlob(dir + "/*.tmpl")
	if err != nil {
		return Settings{}, err
	}

	pref.Content.Templates = tmpl
	return pref, nil
}

// Settings represents a utility struct for passing certain
// properties of a bot around and is required to make bots.
type Settings struct {
	// Telegram API Url
	URL string `json:"url,omitempty"`

	// Telegram token
	Token string `json:"token,omitempty"`

	// Updates channel capacity
	Updates int `json:"updates,omitempty"` // Default: 100

	// Poller is the provider of Updates.
	Poller Poller

	// Reporter is a callback function that will get called
	// on any panics recovered from endpoint handlers.
	Reporter func(error)

	// HTTP Client used to make requests to telegram api
	Client *http.Client

	// You should specify Content's fields in your json config file.
	Content *Content
}

func (pref *Settings) UnmarshalJSON(data []byte) error {
	type SettingsJSON Settings

	var aux struct {
		SettingsJSON
		Webhook    *Webhook    `json:"webhook"`
		LongPoller *LongPoller `json:"long_poller"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	*pref = Settings(aux.SettingsJSON)

	if aux.Webhook != nil {
		pref.Poller = aux.Webhook
	} else if aux.LongPoller != nil {
		pref.Poller = aux.LongPoller
	}

	return nil
}
