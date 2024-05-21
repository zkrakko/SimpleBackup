package gui

import (
	"image/color"
	"log"
	"strings"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Logs struct {
	logNotify chan string
	logLines  []string
	logList   widget.List
	window    *app.Window
	theme     *material.Theme
}

func NewLogs(window *app.Window, theme *material.Theme) *Logs {
	logs := &Logs{
		logNotify: make(chan string, 100),
		window:    window,
		theme:     theme,
	}
	logs.logList.List.Axis = layout.Vertical
	logs.logList.ScrollToEnd = true
	log.SetOutput(logs)
	log.SetFlags(0)
	return logs
}

func (l *Logs) Write(p []byte) (n int, err error) {
	select {
	case l.logNotify <- strings.TrimRight(string(p), "\n\t"):
	default:
	}
	return len(p), nil
}

func (l *Logs) LogNotify() chan string {
	return l.logNotify
}

func (l *Logs) LogReceived(line string) {
	l.logLines = append(l.logLines, line)
	if len(l.logLines) > 1000 {
		l.logLines = l.logLines[len(l.logLines)-1000:]
	}
	l.window.Invalidate()
}

func (l *Logs) Layout(gtx C) D {
	return layout.Inset{Top: 25, Bottom: 25, Left: 10, Right: 10}.Layout(gtx, func(gtx C) D {
		return material.List(l.theme, &l.logList).Layout(gtx, len(l.logLines), func(gtx C, i int) D {
			logLine := l.logLines[i]
			label := material.Body1(l.theme, logLine)
			if strings.HasPrefix(logLine, "ERROR") {
				label.Color = color.NRGBA{R: 192, G: 0, B: 0, A: 255}
				label.Font.Weight = font.Bold
			}
			return label.Layout(gtx)
		})
	})
}
