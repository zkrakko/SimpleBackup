package ftp

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/zkrakko/SimpleBackup/connector"
	"github.com/zkrakko/SimpleBackup/utils"
)

type config struct {
	Ftp struct {
		Server   string `json:"server" validate:"required"`
		User     string `json:"user" validate:"required"`
		Password string `json:"password" validate:"required"`
	} `json:"ftp"`
}

type FTPConnector struct {
	cfg  config
	conn *ftp.ServerConn
}

func New(configParser *utils.ConfigParser) (*FTPConnector, error) {
	cfg := config{}
	err := configParser.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &FTPConnector{cfg: cfg}, nil
}

func (c *FTPConnector) Connect() error {
	log.Printf("connecting to %s@%s", c.cfg.Ftp.User, c.cfg.Ftp.Server)
	conn, err := ftp.Dial(c.cfg.Ftp.Server, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	err = conn.Login(c.cfg.Ftp.User, c.cfg.Ftp.Password)
	if err != nil {
		conn.Quit()
		return err
	}
	c.conn = conn
	return nil
}

func (c *FTPConnector) Disconnect() error {
	log.Print("disconnecting from server")
	err := c.conn.Quit()
	if err != nil {
		return err
	}
	c.conn = nil
	return nil
}

func (c *FTPConnector) MkDirs(path string) error {
	err := c.conn.MakeDir(path)
	if err != nil && filepath.Dir(path) != path {
		err := c.MkDirs(filepath.Dir(path))
		if err != nil {
			return err
		}
		return c.conn.MakeDir(path)
	}
	return err
}

func (c *FTPConnector) Upload(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.conn.Stor(dst, f)
}

func (c *FTPConnector) GetFileInfo(path string) (*connector.FileInfo, error) {
	entry, err := c.conn.GetEntry(path)
	if err != nil {
		return nil, err
	}
	return &connector.FileInfo{Name: entry.Name, Size: int64(entry.Size), Time: entry.Time}, nil
}
