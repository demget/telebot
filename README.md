
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
    pref, err := tb.NewSettings("bot.json", "data")
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

    b.Handle("/start", hander.OnStart)
    b.Handle("/item", hander.OnItem)
    b.Handle(b.InlineButton("refresh"), hander.OnRefresh)
    b.Handle(b.InlineButton("remove"), hander.OnRemove)

    b.Start()
}
```

## Texts
Put all messages' texts in another folder, e.g. `data`. Each message is `*.tmpl` file that will be parsed and executed by [`text/template`](https://golang.org/pkg/text/template) engine.

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

## Inline keyboards + strings
```json
{
    "strings": {
        "error": "Error! Try again later"
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
    /* refresh */
    b.Respond(c)
}

func OnRemove(c *tb.Callback) {
    /* remove */
    if err != nil {
        b.Respond(c, &tb.CallbackResponse{Text: b.String("error")})
    }
}
```

And very simple handling:
```go
b.Handle("/item", hander.OnItem)
b.Handle(b.InlineButton("refresh"), hander.OnRefresh)
b.Handle(b.InlineButton("remove"), hander.OnRemove)
```