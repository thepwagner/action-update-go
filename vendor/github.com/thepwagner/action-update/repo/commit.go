package repo

import (
	"fmt"
	"strings"

	"github.com/thepwagner/action-update/updater"
)

type commitMessageGen func(...updater.Update) string

var defaultCommitMessage = func(updates ...updater.Update) string {
	if len(updates) == 1 {
		update := updates[0]
		return fmt.Sprintf("%s@%s", update.Path, update.Next)
	}
	var s strings.Builder
	s.WriteString("dependency updates\n\n")
	for _, u := range updates {
		_, _ = fmt.Fprintf(&s, "%s@%s\n", u.Path, u.Next)
	}
	return s.String()
}
