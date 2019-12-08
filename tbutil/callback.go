package tbutil

import (
	tb "github.com/demget/telebot"
)

// IsSpam returns true, if passed callback
// contains similar to previous, data.
func IsSpam(c *tb.Callback) bool {
	user, ok := Users.Get(c.Sender.ID)
	if !ok {
		user = Users.Set(c.Sender.ID)
	}

	if user.CallbackData == c.Data {
		return true
	}

	user.CallbackData = c.Data
	return false
}
