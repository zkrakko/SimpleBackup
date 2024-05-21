package synchronizer

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/zkrakko/SimpleBackup/connector"
	"github.com/zkrakko/SimpleBackup/connector/ftp"
	"github.com/zkrakko/SimpleBackup/utils"
)

type config struct {
	Sync struct {
		Folders []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"folders"`
	} `json:"sync"`
}

type Synchronizer struct {
	cfg        config
	connector  connector.Connector
	running    bool
	noProgress bool
	statNotify chan Statistics
}

func New(configParser *utils.ConfigParser, noProgress bool) (*Synchronizer, error) {
	cfg := config{}
	err := configParser.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	conn, err := createConnector(configParser)
	if err != nil {
		return nil, err
	}
	return &Synchronizer{
		cfg:        cfg,
		connector:  conn,
		noProgress: noProgress,
		statNotify: make(chan Statistics),
	}, nil
}

func createConnector(configParser *utils.ConfigParser) (connector.Connector, error) {
	// for now only FTP connector is supported
	conn, err := ftp.New(configParser)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize connector")
	}
	return conn, nil
}

func (s *Synchronizer) StatNotify() chan Statistics {
	return s.statNotify
}

func (s *Synchronizer) GetFolders() []string {
	folders := []string{}
	for _, folder := range s.cfg.Sync.Folders {
		folders = append(folders, folder.From)
	}
	return folders
}

func (s *Synchronizer) Sync() {
	s.running = true
	defer func() {
		s.running = false
	}()

	totalFiles := uint64(0)
	if !s.noProgress {
		totalFiles = s.countFiles()
	}
	stats := newStats(totalFiles, s.statNotify, !s.noProgress)

	errorOccured := false
	err := s.connector.Connect()
	if err != nil {
		log.Printf("ERROR: cannot connect to server: %s", err.Error())
		return
	}
	defer func() {
		s.connector.Disconnect()
		log.Printf("uploaded %d files", stats.UploadedFiles)
		log.Print("backup finished")
	}()

	for _, folder := range s.cfg.Sync.Folders {
		log.Printf("backing up %s to %s", folder.From, folder.To)
		err := s.ensureRemoteDir(folder.To)
		if err != nil {
			log.Printf("ERROR: could not create remote folder: %s", err.Error())
			errorOccured = true
			continue
		}
		err = s.syncDir(folder.From, folder.To, stats)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			errorOccured = true
		}
	}
	if errorOccured {
		log.Printf("ERROR: could not sync all folders")
	}
}

func (s *Synchronizer) IsRunning() bool {
	return s.running
}

func (s *Synchronizer) ensureRemoteDir(path string) error {
	_, err := s.connector.GetFileInfo(path)
	if err != nil {
		return s.connector.MkDirs(path)
	}
	return nil
}

func (s *Synchronizer) syncDir(src, dst string, stats *Statistics) error {
	return fs.WalkDir(os.DirFS(src), ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil || path == "." {
			return nil
		}
		localPath := filepath.Join(src, path)
		remotePath := filepath.Join(dst, path)
		remoteInfo, err := s.connector.GetFileInfo(remotePath)
		// we assume err here means file not exists
		// if there was some other error, consequent operations will fail anyway
		fileNotExists := err != nil
		if dirEntry.IsDir() {
			if fileNotExists {
				err := s.connector.MkDirs(remotePath)
				if err != nil {
					return err
				}
			}
		} else {
			localInfo, err := dirEntry.Info()
			if err != nil {
				return err
			}
			if fileNotExists || s.fileOutdated(localInfo, remoteInfo) {
				log.Printf("backup file %s", localPath)
				err := s.connector.Upload(localPath, remotePath)
				if err != nil {
					return err
				}
				stats.Uploaded()
			}
			stats.Processed()
		}
		return nil
	})
}

func (s *Synchronizer) fileOutdated(localInfo fs.FileInfo, remoteInfo *connector.FileInfo) bool {
	return localInfo.ModTime().Unix() > remoteInfo.Time.Unix() || localInfo.Size() != remoteInfo.Size
}

func (s *Synchronizer) countFiles() uint64 {
	log.Print("counting files")
	totalFiles := uint64(0)
	for _, folder := range s.cfg.Sync.Folders {
		fs.WalkDir(os.DirFS(folder.From), ".", func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil || path == "." {
				return nil
			}
			if !dirEntry.IsDir() {
				totalFiles++
			}
			return nil
		})
	}
	return totalFiles
}
