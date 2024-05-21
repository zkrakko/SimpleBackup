package cli

import (
	"log"

	"github.com/zkrakko/SimpleBackup/synchronizer"
	"github.com/zkrakko/SimpleBackup/utils"
)

type SimpleBackupCli struct {
	synchronizer *synchronizer.Synchronizer
}

func New(configParser *utils.ConfigParser) *SimpleBackupCli {
	synchronizer, err := synchronizer.New(configParser, true)
	if err != nil {
		log.Fatalf("ERROR: cannot initialize synchronizer: %s", err.Error())
	}
	return &SimpleBackupCli{
		synchronizer: synchronizer,
	}
}

func (b *SimpleBackupCli) Run() {
	b.synchronizer.Sync()
}
