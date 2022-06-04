package cuslog

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
	IgnoreBasicFields bool
}

func (t *TextFormatter) Format(e *Entry) error {
	if !t.IgnoreBasicFields {
		e.Buffer.WriteString(fmt.Sprintf("%s %s->", e.Time.Format(time.RFC3339), LevelNameMapping[e.Level]))
		if e.File != "" {
			short := e.File[strings.LastIndex(e.File, "/")+1:]
			e.Buffer.WriteString(fmt.Sprintf("%s:%d", short, e.Line))
		}
		e.Buffer.WriteString(" ")
	}
	switch e.Format {
	case FmtEmptySeparate:
		e.Buffer.WriteString(fmt.Sprint(e.Args...))
	default:
		e.Buffer.WriteString(fmt.Sprintf(e.Format, e.Args...))
	}
	e.Buffer.WriteString("\n")

	return nil
}
