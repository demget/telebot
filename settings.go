package telebot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
)

// TemplateFuncMap is pre-defined functions that can be used in your templates.
var TemplateFuncMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },

	// Escapes double-quotes. Useful in json templates.
	"jsq": func(s string) string {
		return strings.ReplaceAll(s, `"`, `\"`)
	},
}

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

	tmpl, err := template.New("data").
		Funcs(TemplateFuncMap).
		ParseGlob(dir + "/*.tmpl")
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

		Strings       map[string]string          `json:"strings"`
		InlineButtons map[string]json.RawMessage `json:"inline_buttons"`
		InlineResults map[string]json.RawMessage `json:"inline_results"`
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

	cont := &Content{
		Strings:       template.New("strings").Funcs(TemplateFuncMap),
		InlineButtons: template.New("inline_buttons").Funcs(TemplateFuncMap),
		InlineResults: template.New("inline_results").Funcs(TemplateFuncMap),
	}

	for k, v := range aux.Strings {
		_, err := cont.Strings.New(k).Parse(v)
		if err != nil {
			return err
		}
	}
	for k, v := range aux.InlineButtons {
		_, err := cont.InlineButtons.New(k).Parse(string(v))
		if err != nil {
			return err
		}
	}
	for k, v := range aux.InlineResults {
		_, err := cont.InlineResults.New(k).Parse(string(v))
		if err != nil {
			return err
		}
	}

	pref.Content = cont
	return nil
}
