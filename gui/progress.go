package gui

import (
	"gioui.org/app"
	"gioui.org/widget/material"
	"github.com/zkrakko/SimpleBackup/synchronizer"
)

type Progress struct {
	window *app.Window
	theme  *material.Theme
	stats  synchronizer.Statistics
}

func NewProgress(window *app.Window, theme *material.Theme) *Progress {
	return &Progress{
		window: window,
		theme:  theme,
	}
}
func (p *Progress) StatsReceived(stats synchronizer.Statistics) {
	p.stats = stats
	p.window.Invalidate()
}

func (p *Progress) Layout(gtx C) D {
	return material.ProgressBar(p.theme, p.stats.Progress()).Layout(gtx)
}
