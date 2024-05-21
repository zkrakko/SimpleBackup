package gui

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/zkrakko/SimpleBackup/synchronizer"
	"github.com/zkrakko/SimpleBackup/utils"
)

type C = layout.Context
type D = layout.Dimensions

type SimpleBackupGui struct {
	synchronizer *synchronizer.Synchronizer
}

func New(configParser *utils.ConfigParser, noProgress bool) *SimpleBackupGui {
	synchronizer, err := synchronizer.New(configParser, noProgress)
	if err != nil {
		log.Fatalf("ERROR: cannot initialize synchronizer: %s", err.Error())
	}
	return &SimpleBackupGui{
		synchronizer: synchronizer,
	}
}

func (g *SimpleBackupGui) Run() {
	go GuiMain(g.synchronizer)
	app.Main()
}

type GuiContext struct {
	synchronizer *synchronizer.Synchronizer
	events       chan event.Event
	acks         chan struct{}
	ops          op.Ops
	window       *app.Window
	theme        *material.Theme
	logs         *Logs
	progress     *Progress
	backupButton *BackupButton
}

func GuiMain(synchronizer *synchronizer.Synchronizer) {
	window := createWindow()
	theme := createTheme()
	context := GuiContext{
		synchronizer: synchronizer,
		events:       make(chan event.Event),
		acks:         make(chan struct{}),
		window:       window,
		theme:        theme,
		logs:         NewLogs(window, theme),
		progress:     NewProgress(window, theme),
		backupButton: NewBackupButton(theme, synchronizer),
	}
	log.Print("the following folders will be backed up:")
	for _, folder := range synchronizer.GetFolders() {
		log.Printf("    %s", folder)
	}
	go context.consumeEvents()
	err := context.eventLoop()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func createTheme() *material.Theme {
	theme := material.NewTheme()
	theme.TextSize = 14
	return theme
}

func createWindow() *app.Window {
	window := new(app.Window)
	window.Option(
		app.Size(500, 500),
		app.MinSize(500, 300),
		app.Title("Simple Backup"),
	)
	return window
}

func (c *GuiContext) eventLoop() error {
	for {
		select {
		case line := <-c.logs.LogNotify():
			c.logs.LogReceived(line)
		case stats := <-c.synchronizer.StatNotify():
			c.progress.StatsReceived(stats)
		case e := <-c.events:
			switch e := e.(type) {
			case app.DestroyEvent:
				c.acks <- struct{}{}
				return e.Err
			case app.FrameEvent:
				c.frameEvent(e)
			}
			c.acks <- struct{}{}
		}
	}
}

func (c *GuiContext) consumeEvents() {
	for {
		ev := c.window.Event()
		c.events <- ev
		<-c.acks
		if _, ok := ev.(app.DestroyEvent); ok {
			return
		}
	}
}

func (c *GuiContext) frameEvent(event app.FrameEvent) {
	gtx := app.NewContext(&c.ops, event)
	if c.backupButton.Clicked(gtx) {
		log.Print("")
		go c.synchronizer.Sync()
	}
	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return c.progress.Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return c.logs.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return c.backupButton.Layout(gtx)
		}),
	)
	event.Frame(gtx.Ops)
}
