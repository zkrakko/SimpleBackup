package gui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/zkrakko/SimpleBackup/synchronizer"
)

type BackupButton struct {
	theme        *material.Theme
	button       widget.Clickable
	synchronizer *synchronizer.Synchronizer
}

func NewBackupButton(theme *material.Theme, synchronizer *synchronizer.Synchronizer) *BackupButton {
	return &BackupButton{
		theme:        theme,
		synchronizer: synchronizer,
	}
}

func (b *BackupButton) Clicked(gtx C) bool {
	return b.button.Clicked(gtx) && !b.synchronizer.IsRunning()
}

func (b *BackupButton) Layout(gtx C) D {
	return layout.Inset{Top: 0, Bottom: 25, Right: 100, Left: 100}.Layout(gtx, func(gtx C) D {
		text := "Backup"
		if b.synchronizer.IsRunning() {
			gtx = gtx.Disabled()
			text = "Working..."
		}
		return material.Button(b.theme, &b.button, text).Layout(gtx)
	})
}
