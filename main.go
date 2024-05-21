package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/zkrakko/SimpleBackup/cli"
	"github.com/zkrakko/SimpleBackup/gui"
	"github.com/zkrakko/SimpleBackup/utils"
)

func createConfigParser() *utils.ConfigParser {
	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("ERROR: cannot get executable: %s", err.Error())
	}
	exePath := filepath.Dir(ex)
	configParser, err := utils.NewConfigParser(filepath.Join(exePath, "config.yaml"))
	if err != nil {
		log.Fatalf("ERROR: cannot create config parser: %s", err.Error())
	}
	return configParser
}

func main() {
	useGui := flag.Bool("gui", false, "If specified, started with GUI")
	noProgress := flag.Bool("noprogress", false, "If specified, progress is not displayed (speed up in case of lot of files)")
	flag.Parse()
	configParser := createConfigParser()
	if *useGui {
		backup := gui.New(configParser, *noProgress)
		backup.Run()
	} else {
		backup := cli.New(configParser)
		backup.Run()
	}
}
