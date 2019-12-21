package telebot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ghodss/yaml"
)

// NewSettings does try to load Settings from your json config file.
// 	- path is config path
// 	- tmplEngine is implementation of the templating engine
func NewSettings(path string, tmplEngine Template) (Settings, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}
	return newSettings(data, tmplEngine)
}

// NewSettingsYAML does try to load Settings from your yaml config file.
// 	- path is config path
// 	- tmplEngine is implementation of the templating engine
func NewSettingsYAML(path string, tmplEngine Template) (Settings, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}
	data, err = yaml.YAMLToJSON(data)
	if err != nil {
		return Settings{}, err
	}
	return newSettings(data, tmplEngine)
}

func newSettings(data []byte, tmplEngine Template) (Settings, error) {
	var pref Settings
	pref.TemplateEngine = tmplEngine

	if err := json.Unmarshal(data, &pref); err != nil {
		return Settings{}, err
	}
	if err := json.Unmarshal(data, &pref.Content); err != nil {
		return Settings{}, err
	}

	tmpl := tmplEngine.New("data")
	if err := tmpl.ParseGlob(); err != nil {
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

	// Passed template engine, that will be used for all executable content.
	TemplateEngine Template

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

	aux.TemplateEngine = pref.TemplateEngine
	*pref = Settings(aux.SettingsJSON)

	if aux.Webhook != nil {
		pref.Poller = aux.Webhook
	} else if aux.LongPoller != nil {
		pref.Poller = aux.LongPoller
	}

	cont := &Content{
		Strings:       pref.TemplateEngine.New("strings"),
		InlineButtons: pref.TemplateEngine.New("inline_buttons"),
		InlineResults: pref.TemplateEngine.New("inline_results"),
	}

	for k, v := range aux.Strings {
		err := cont.Strings.Parse(k, v)
		if err != nil {
			return err
		}
	}
	for k, v := range aux.InlineButtons {
		err := cont.InlineButtons.Parse(k, string(v))
		if err != nil {
			return err
		}
	}
	for k, v := range aux.InlineResults {
		err := cont.InlineResults.Parse(k, string(v))
		if err != nil {
			return err
		}
	}

	pref.Content = cont
	return nil
}
