# Telebot
>"I never knew creating Telegram bots could be so _sexy_!"

[![GoDoc](https://godoc.org/gopkg.in/tucnak/telebot.v2?status.svg)](https://godoc.org/gopkg.in/tucnak/telebot.v2)
[![Travis](https://travis-ci.org/tucnak/telebot.svg?branch=v2)](https://travis-ci.org/tucnak/telebot)

```bash
go get -u gopkg.in/tucnak/telebot.v2
```

* [Overview](#overview)
* [Getting Started](#getting-started)
	- [Poller](#poller)
	- [Commands](#commands)
	- [Files](#files)
	- [Sendable](#sendable)
	- [Editable](#editable)
	- [Keyboards](#keyboards)
	- [Inline mode](#inline-mode)
* [Contributing](#contributing)
* [Donate](#donate)
* [License](#license)

# Overview
Telebot is a bot framework for [Telegram](https://telegram.org) [Bot API](https://core.telegram.org/bots/api).
This package provides the best of its kind API for command routing, inline query requests and keyboards, as well
as callbacks. Actually, I went a couple steps further, so instead of making a 1:1 API wrapper I chose to focus on
the beauty of API and performance. Some of the strong sides of telebot are:

* Real concise API
* Command routing
* Middleware
* Transparent File API
* Effortless bot callbacks

All the methods of telebot API are _extremely_ easy to memorize and get used to. Also, consider Telebot a
highload-ready solution. I'll test and benchmark the most popular actions and if necessary, optimize
against them without sacrificing API quality.

# Getting Started
Let's take a look at the minimal telebot setup:
```go
package main

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	b, err := tb.NewBot(tb.Settings{
		Token:  "TOKEN_HERE",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, "hello world")
	})

	b.Start()
}

```

Simple, innit? Telebot's routing system takes care of deliviering updates
to their endpoints, so in order to get to handle any meaningful event,
all you got to do is just plug your function to one of the Telebot-provided
endpoints. You can find the full list
[here](https://godoc.org/gopkg.in/tucnak/telebot.v2#pkg-constants).

```go
b, _ := tb.NewBot(settings)

b.Handle(tb.OnText, func(m *tb.Message) {
	// all the text messages that weren't
	// captured by existing handlers
}

b.Handle(tb.OnPhoto, func(m *tb.Message) {
	// photos only
}

b.Handle(tb.OnChannelPost, func (m *tb.Message) {
	// channel posts only
})

b.Handle(tb.Query, func (q *tb.Query) {
	// incoming inline queries
})
```

Now there's a dozen of supported endpoints (see package consts). Let me know
if you'd like to see some endpoint or endpoint idea implemented. This system
is completely extensible, so I can introduce them without braking
backwards-compatibity.

## Poller
Telebot doesn't really care how you provide it with incoming updates, as long
as you set it up with a Poller:
```go
// Poller is a provider of Updates.
//
// All pollers must implement Poll(), which accepts bot
// pointer and subscription channel and start polling
// synchronously straight away.
type Poller interface {
	// Poll is supposed to take the bot object
	// subscription channel and start polling
	// for Updates immediately.
	//
	// Poller must listen for stop constantly and close
	// it as soon as it's done polling.
	Poll(b *Bot, updates chan Update, stop chan struct{})
}
```

Telegram Bot API supports long polling and webhook integration. I don't really
care about webhooks, so the only concrete Poller you'll find in the library
is the `LongPoller`. Poller means you can plug telebot into whatever existing
bot infrastructure (load balancers?) you need, if you need to. Another great thing
about pollers is that you can chain them, making some sort of middleware:
```go
poller := &tb.LongPoller{Timeout: 15 * time.Second}
spamProtected := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
	if upd.Message == nil {
		return true
	}

	if strings.Contains(upd.Message.Text, "spam") {
		return false
	}

	return true
})

bot, _ := tb.NewBot(tb.Settings{
	// ...
	Poller: spamProtected,
})

// graceful shutdown
go func() {
	<-time.After(N * time.Second)
	bot.Stop()
})()

bot.Start() // blocks until shutdown

fmt.Println(poller.LastUpdateID) // 134237
```

## Commands
When handling commands, Telebot supports both direct (`/command`) and group-like
syntax (`/command@botname`) and will never deliver messages addressed to some
other bot, even if [privacy mode](https://core.telegram.org/bots#privacy-mode) is off.
For simplified deep-linking, telebot also extracts payload:
```go
// Command: /start <PAYLOAD>
b.Handle("/start", func(m *tb.Message) {
	if !m.Private() {
		return
	}

	fmt.Println(m.Payload) // <PAYLOAD>
})
```

## Files
>Telegram allows files up to 20 MB in size.

Telebot allows to both upload (from disk / by URL) and download (from Telegram)
and files in bot's scope. Also, sending any kind of media with a File created
from disk will upload the file to Telegram automatically:
```go
a := &tb.Audio{File: tb.FromDisk("file.ogg")}

fmt.Println(a.OnDisk()) // true
fmt.Println(a.InCloud()) // false

// Will upload the file from disk and send it to recipient
bot.Send(recipient, a)

// Next time you'll be sending this very *Audio, Telebot won't
// re-upload the same file but rather utilize its Telegram FileID
bot.Send(otherRecipient, a)

fmt.Println(a.OnDisk()) // true
fmt.Println(a.InCloud()) // true
fmt.Println(a.FileID) // <telegram file id: ABC-DEF1234ghIkl-zyx57W2v1u123ew11>
```

You might want to save certain `File`s in order to avoid re-uploading. Feel free
to marshal them into whatever format, `File` only contain public fields, so no
data will ever be lost.

## Sendable
Send is undoubteldy the most important method in Telebot. `Send()` accepts a
`Recipient` (could be user, group or a channel) and a `Sendable`. FYI, not only
all telebot-provided media types (`Photo`, `Audio`, `Video`, etc.) are `Sendable`,
but you can create composite types of your own. As long as they satisfy `Sendable`,
Telebot will be able to send them out.

```go
// Sendable is any object that can send itself.
//
// This is pretty cool, since it lets bots implement
// custom Sendables for complex kind of media or
// chat objects spanning across multiple messages.
type Sendable interface {
    Send(*Bot, Recipient, *SendOptions) (*Message, error)
}
```

The only type at the time that doesn't fit `Send()` is `Album` and there is a reason
for that. Albums were added not so long ago, so they are slightly quirky for backwards
compatibilities sake. In fact, an `Album` can be sent, but never received. Instead,
Telegram returns a `[]Message`, one for each media object in the album:
```go
p := &tb.Photo{File: tb.FromDisk("chicken.jpg")}
v := &tb.Video{File: tb.FromURL("http://video.mp4")}

msgs, err := b.SendAlbum(user, tb.Album{p, v})
```

### Send options
Send options are objects and flags you can pass to `Send()`, `Edit()` and friends
as optional arguments (following the recipient and the text/media). The most
important one is called `SendOptions`, it lets you control _all_ the properties of
the message supported by Telegram. The only drawback is that it's rather
inconvenient to use at times, so `Send()` supports multiple shorthands:
```go
// regular send options
b.Send(user, "text", &tb.SendOptions{
	// ...
})

// ReplyMarkup is a part of SendOptions,
// but often it's the only option you need
b.Send(user, "text", &tb.ReplyMarkup{
	// ...
})

// flags: no notification && no web link preview
b.Send(user, "text", tb.Silent, tb.NoPreview)
```

Full list of supported option-flags you can find
[here](https://github.com/tucnak/telebot/blob/v2/options.go#L9).

## Editable
If you want to edit some existing message, you don't really need to store the
original `*Message` object. In fact, upon edit, Telegram only requires two IDs:
ChatID and MessageID. And it doesn't really require the whole Message. Also you
might want to store references to certain messages in the database, so for me it
made sense for *any* Go struct to be editable as Telegram message, to implement
Editable interface:
```go
// Editable is an interface for all objects that
// provide "message signature", a pair of 32-bit
// message ID and 64-bit chat ID, both required
// for edit operations.
//
// Use case: DB model struct for messages to-be
// edited with, say two collums: msg_id,chat_id
// could easily implement MessageSig() making
// instances of stored messages editable.
type Editable interface {
	// MessageSig is a "message signature".
	//
	// For inline messages, return chatID = 0.
	MessageSig() (messageID int, chatID int64)
}
```

For example, `Message` type is Editable. Here is an implementation of `StoredMessage`
type, provided by telebot:
```go
// StoredMessage is an example struct suitable for being
// stored in the database as-is or being embedded into
// a larger struct, which is often the case (you might
// want to store some metadata alongside, or might not.)
type StoredMessage struct {
	MessageID int   `sql:"message_id" json:"message_id"`
	ChatID    int64 `sql:"chat_id" json:"chat_id"`
}

func (x StoredMessage) MessageSig() (int, int64) {
	return x.MessageID, x.ChatID
}
```

Why bother at all? Well, it allows you to do things like this:
```go
// just two integer columns in the database
var msgs []tb.StoredMessage
db.Find(&msgs) // gorm syntax

for _, msg := range msgs {
	bot.Edit(&msg, "Updated text.")
	// or
	bot.Delete(&msg)
}
```

I find it incredibly neat. Worth noting, at this point of time there exists
another method in the Edit family, `EditCaption()` which is of a pretty
rare use, so I didn't bother including it to `Edit()`, just like I did with
`SendAlbum()` as it would inevitably lead to unnecessary complications.
```go
var m *Message

// change caption of a photo, audio, etc.
bot.EditCaption(m, "new caption")
```

## Keyboards
Telebot supports both kinds of keyboards Telegram provides: reply and inline
keyboards. All buttons can act as endpoints for `Handle()`:
`Handle()`:

```go
func main() {
	b, _ := tb.NewBot(tb.Settings{...})

	// This button will be displayed in user's
	// reply keyboard.
	replyBtn := tb.ReplyButton{Text: "🌕 Button #1"}
	replyKeys := [][]tb.ReplyButton{
		[]tb.ReplyButton{replyBtn},
		// ...
	}

	// And this one — just under the message itself.
	// Pressing it will cause the client to send
	// the bot a callback.
	//
	// Make sure Unique stays unique as it has to be
	// for callback routing to work.
	inlineBtn := tb.InlineButton{
		Unique: "sad_moon",
		Text: "🌚 Button #2",
	}
	inlineKeys := [][]tb.InlineButton{
		[]tb.InlineButton{inlineBtn},
		// ...
	}

	b.Handle(&replyBtn, func(m *tb.Message) {
		// on reply button pressed
	})

	b.Handle(&inlineBtn, func(c *tb.Callback) {
		// on inline button pressed (callback!)

		// always respond!
		c.Respond(&tb.CallbackResponse{...})
	})

	// Command: /start <PAYLOAD>
	b.Handle("/start", func(m *tb.Message) {
		if !m.Private() {
			return
		}

		b.Send(m.Sender, "Hello!", &tb.ReplyMarkup{
			ReplyKeyboard:  replyKeys,
			InlineKeyboard: inlineKeys,
		})
	})

	b.Start()
}
```

## Inline mode
So if you want to handle incoming inline queries you better plug the `tb.OnQuery`
endpoint and then use the `Answer()` method to send a list of inline queries
back. I think at the time of writing, telebot supports all of the provided result
types (but not the cached ones). This is how it looks like:

```go
b.Handle(tb.OnQuery, func(q *tb.Query) {
	urls := []string{
		"http://photo.jpg",
		"http://photo2.jpg",
	}

	results := make(tb.Results, len(urls)) // []tb.Result
	for i, url := range urls {
		result := &tb.PhotoResult{
			URL: url,

			// required for photos
			ThumbURL: url,
		}

		results[i] = result
	}

	err := b.Answer(q, &tb.QueryResponse{
		Results: results,
		CacheTime: 60, // a minute
	})

	if err != nil {
		fmt.Println(err)
	}
})
```

There's not much to talk about really. It also support some form of authentication
through deep-linking. For that, use fields `SwitchPMText` and `SwitchPMParameter`
of `QueryResponse`.

# Contributing

1. Fork it
2. Clone it: `git clone https://github.com/tucnak/telebot`
3. Create your feature branch: `git checkout -b my-new-feature`
4. Make changes and add them: `git add .`
5. Commit: `git commit -m 'Add some feature'`
6. Push: `git push origin my-new-feature`
7. Pull request

# Donate

I do coding for fun but I also try to search for interesting solutions and
optimize them as much as possible.
If you feel like it's a good piece of software, I wouldn't mind a tip!

Bitcoin: `1DkfrFvSRqgBnBuxv9BzAz83dqur5zrdTH`

# License

Telebot is distributed under MIT.
