package connector

import "time"

type FileInfo struct {
	Name string
	Size int64
	Time time.Time
}
type Connector interface {
	Connect() error
	Disconnect() error
	MkDirs(path string) error
	Upload(src, dst string) error
	GetFileInfo(path string) (*FileInfo, error)
}
