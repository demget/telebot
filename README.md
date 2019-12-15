
# What is `content` branch

This fork is simple and very helpful if you develop big and complex projects. The main goal is to manage bot's content easily – texts, buttons, keyboards etc. Also this package contains some patches and fixes by me that is still not accepted in original [`tucnak/telebot`](https://github.com/tucnak/telebot) repository. More bot examples will be added in the future.

**Feel free to use, contribute and ask questions!**

# How to use 

```go
package main

import tb "github.com/demget/telebot"

func main() {
	// "bot.json" is your config file
	// "data" is your texts directory
	pref, err := tb.NewSettings("bot.json", &tb.TemplateText{
	    Dir: "data",
    })
	if err != nil {
		log.Fatalln(err)
	}

	// you also can save token in bot.json
	pref.Token = os.Getenv("TOKEN") 
	pref.Reporter = report

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatalln(err)
	}

	b.Handle("/start", handler.OnStart)
	b.Handle("/item", handler.OnItem)
	b.Handle(b.InlineButton("refresh"), handler.OnRefresh)
	b.Handle(b.InlineButton("remove"), handler.OnRemove)

	b.Start()
}
```

## Texts
Put all messages' texts in another folder, e.g. `data`. Each message is `*.tmpl` file that will be parsed and executed by [`text/template`](https://golang.org/pkg/text/template) engine or by [`aymerick/raymond`](https://github.com/aymerick/raymond).
You can to select between them:
* `tb.TemplateText` implements the `text/template` library
* `tb.TemplateHandlebars` implements the `aymerick/raymond` library
### Example
```go
func OnStart(m *tb.Message) {
	b.Send(m.Sender, b.Text("hello", m.Sender), tb.ModeMarkdown)
}
```

> `Hi, *{{.FirstName}}*!` → Hi, **Pavel**!

> `Hi, {{if .Username}}@{{.Username}}{{else}}*{{.FirstName}}*{{end}}!` → Hi, [@durov]()!

## Vars
```json
{
	"vars": {
		"secret": "qz_BuGo2",
		"admins": [],
		"limits": {
			"max_requests_per_user": 20
		}
	},
}
```
```go
package app

type Config struct {
	Secret string `json:"secret"`
	Admins []int  `json:"admins"`
	Limits struct {
		MaxRequestsPerUser int `json:"max_requests_per_user"`
		// ...
	}
}
```
```go
var conf app.Config
if err := b.Vars(&conf); err != nil {
	log.Fatalln(err)
}

// now you can use your variables:
	conf.Secret
	conf.Admins
	conf.Limits
```

## Reply keyboards
```json
{
	"buttons": {
		"help": "❓ Help",
		"settings": "⚙️ Settings"
	},
	"keyboards": {
		"menu": [["help", "settings"]]
	},
}
```
```go
func OnStart(m *tb.Message) {
	b.Send(m.Sender, 
		b.Text("hello", m.Sender), 
		b.Markup("menu"),
		tb.ModeMarkdown)
}
```

## Inline keyboards + Strings
```json
{
	"strings": {
		"removed": "Removed successfully!"
	},
	"inline_buttons": {
		"refresh": {
			"unique": "refresh",
			"callback_data": "{{.ID}}",
			"text": "🔄 Refresh"
		},
		"remove": {
			"unique": "remove",
			"callback_data": "{{.ID}}",
			"text": "🛑 Remove"
		}
	},
	"inline_keyboards": {
		"item": [["refresh"], ["remove"]],
	},
}
```
```go
package handler

func OnItem(m *tb.Message) {
	b.Send(m.Sender, 
		b.Text("item", item), 
		b.InlineMarkup("item", item),
		tb.ModeMarkdown)
}

func OnRefresh(c *tb.Callback) {
	defer b.Respond(c)
	/* refresh */
}

func OnRemove(c *tb.Callback) {
	/* remove */

	b.Respond(c, &tb.CallbackResponse{
		Text:      b.String("removed"),
		ShowAlert: true,
	})
}
```

And very simple handling:
```go
b.Handle("/item", handler.OnItem)
b.Handle(b.InlineButton("refresh"), handler.OnRefresh)
b.Handle(b.InlineButton("remove"), handler.OnRemove)
```

## Inline query results
```json
{
	"inline_results": {
		"item": {
			"type": "article",
			"id": "{{.ID}}",
			"title": "{{jsq .Title}}",
			"description": "{{jsq .Description}}",
			"thumb_url": "{{.ThumbnailURL}}"
		}
	},
}
```
```go
package handler

func OnQuery(q *tb.OnQuery) {
	var results tb.Results
	for _, item := range items {
		result := b.InlineResult("item", item)
		if result == nil { // something went wrong
			continue
		}

		result.SetContent(&tb.InputTextMessageContent{
			Text:      b.Text("item", item),
			ParseMode: tb.ModeHTML,
		})

		result.SetReplyMarkup(b.InlineMarkup("item", item))
		results = append(results, result)
	}

	b.Answer(q, &tb.QueryResponse{
		Results:   results,
		CacheTime: 300,
	})
}
```

## Additional and custom template functions

There are some additional template functions which are accessible in any text template and config. Some simple things, that standard template package still not do. Check `settings.go` for all pre-defined functions. The list will be extended in the future.

You can add your custom functions before creating bot instance like so:
```go
func init() {
	tb.TemplateFuncMap["upper"] = strings.ToUpper

	tb.TemplateFuncMap["name"] = func(s string) {
		// your custom function
	}
}
```

**Examples:**

> `{{add .N 4}}` → `9`

> `{{sub .N 4}}` → `1`

> ```{{jsq `Some \weird json-incompatible "title"`}}``` → ```Some \\weird json-incompatible \"title\"```
