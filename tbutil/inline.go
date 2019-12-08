package tbutil

import (
	"time"

	tb "github.com/demget/telebot"
)

// LimitQuery helps to limit inline query handling.
// For example, user's input:
// 		@inlinebot V
// 		@inlinebot Ver
// 		@inlinebot Very lo
// 		@inlinebot Very long u
// 		@inlinebot Very long user's
// 		@inlinebot Very long user's query
//
// It waits passed duration before returning true,
// which means that caller can handle this query.
// If in that period came new query, old select statement
// interrupts and new waiting process call again.
//
// Works bad. TODO: Fix and improve.
func LimitQuery(q *tb.Query, d time.Duration) bool {
	user, ok := Users.Get(q.From.ID)
	if !ok {
		user = Users.Set(q.From.ID)
	}

	if user.QueryChan == nil {
		user.QueryChan = make(chan struct{}, 1)
	} else {
		user.QueryChan <- struct{}{}
	}

	select {
	case <-time.After(d):
		user.QueryChan = nil
		break
	case <-user.QueryChan:
		return false
	}
	return true
}
