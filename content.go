package telebot

import (
	"bytes"
	"encoding/json"
	"log"
	"text/template"

	"github.com/pkg/errors"
)

// Content stores bot's buttons, markups, texts etc.
type Content struct {
	// RawVars is raw encoded vars struct
	// that may store your specific Config struct.
	RawVars json.RawMessage `json:"vars"`

	// Strings can be used for storing bot's specific strings
	// that you can use in your messages or alerts.
	// To format the string call bot.String("key", args...).
	Strings *template.Template `json:"-"`

	// Simple ReplyMarkup entities.
	Buttons   map[string]string     `json:"buttons"`
	Keyboards map[string][][]string `json:"keyboards"`

	// InlineMarkup entities.
	InlineButtons   *template.Template    `json:"-"`
	InlineKeyboards map[string][][]string `json:"inline_keyboards"`

	// InilineQuery result entities.
	InlineResults *template.Template `json:"-"`

	// Templates stores all bot's messages – must be valid "text/template"
	// templates with ".tmpl" ext. You should save it as separated files.
	// This field fills automatically when you create settings via NewSettings.
	Templates *template.Template `json:"-"`
}

// Vars parses the raw JSON vars and stores the result
// in the value pointed to by v.
func (c *Content) Vars(v interface{}) error {
	return json.Unmarshal(c.RawVars, v)
}

// Text returns executed template from Templates map.
// It uses "text/template" parser.
func (c *Content) Text(key string, args ...interface{}) string {
	if c.Templates == nil {
		return ""
	}

	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}

	var buf bytes.Buffer
	if err := c.Templates.ExecuteTemplate(&buf, key+".tmpl", arg); err != nil {
		c.debug(err)
	}
	return buf.String()
}

// String returns formatted string from Strings map.
func (c *Content) String(key string, args ...interface{}) string {
	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}

	var buf bytes.Buffer
	if err := c.Strings.ExecuteTemplate(&buf, key, arg); err != nil {
		c.debug(err)
	}
	return buf.String()
}

// Button returns ReplyButton with text from Buttons map.
func (c *Content) Button(key string) *ReplyButton {
	return &ReplyButton{Text: c.Buttons[key]}
}

// Markup returns markup with ReplyKeyboard.
func (c *Content) Markup(key string) *ReplyMarkup {
	keyb, ok := c.Keyboards[key]
	if !ok {
		return nil
	}

	markup := new(ReplyMarkup)
	markup.ReplyKeyboard = make([][]ReplyButton, len(keyb))

	// You can't manage these fields in config file for now.
	// I usually need only ResizeReplyKeyboard option.
	markup.ResizeReplyKeyboard = true
	// markup.OneTimeKeyboard = false
	// markup.Selective = false
	// markup.ForceReply = false

	for i, btns := range keyb {
		var row []ReplyButton
		for _, btn := range btns {
			row = append(row, *c.Button(btn))
		}
		markup.ReplyKeyboard[i] = row
	}

	return markup
}

// InlineButton returns formatted InlineButton.
// It uses "text/template" parser.
func (c *Content) InlineButton(key string, args ...interface{}) *InlineButton {
	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}

	var buf bytes.Buffer
	if err := c.InlineButtons.ExecuteTemplate(&buf, key, arg); err != nil {
		c.debug(err)
		return nil
	}

	var btn InlineButton
	if err := json.Unmarshal(buf.Bytes(), &btn); err != nil {
		c.debug(err)
		return nil
	}
	return &btn
}

// InlineMarkup returns markup with formatted InineKeyboard.
// It uses "text/template" parser.
func (c *Content) InlineMarkup(key string, args ...interface{}) *ReplyMarkup {
	keyb, ok := c.InlineKeyboards[key]
	if !ok {
		return nil
	}

	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}

	markup := new(ReplyMarkup)
	markup.InlineKeyboard = make([][]InlineButton, len(keyb))

	for i, btns := range keyb {
		var row []InlineButton
		for _, btn := range btns {
			row = append(row, *c.InlineButton(btn, arg))
		}
		markup.InlineKeyboard[i] = row
	}

	return markup
}

// InlineResult returns formatted inline query result.
// It uses "text/template" parser.
func (c *Content) InlineResult(key string, args ...interface{}) Result {
	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}

	var buf bytes.Buffer
	if err := c.InlineResults.ExecuteTemplate(&buf, key, arg); err != nil {
		c.debug(err)
		return nil
	}
	bs := buf.Bytes()

	var aux struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(bs, &aux); err != nil {
		c.debug(err)
		return nil
	}

	switch aux.Type {
	case "article":
		var r ArticleResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "audio":
		var r AudioResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "contact":
		var r ContactResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "document":
		var r DocumentResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "gif":
		var r GifResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "location":
		var r LocationResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "mpeg4_gif":
		var r Mpeg4GifResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "photo":
		var r PhotoResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "venue":
		var r VenueResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "video":
		var r VideoResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "voice":
		var r VoiceResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	case "sticker":
		var r StickerResult
		if err := json.Unmarshal(bs, &r); err != nil {
			c.debug(err)
			break
		}
		return &r
	}

	return nil
}

func (c *Content) debug(err error) {
	// it's better to implement global package logger
	err = errors.WithStack(err)
	log.Printf("%+v\n", err)
}
