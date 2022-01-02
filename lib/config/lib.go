package config

import (
	"log"
	"os"
	"os/user"
	"path"

	"github.com/natefinch/lumberjack"
	"gopkg.in/yaml.v2"
)

type cfg struct {
	Addr   string            `json:"addr" yaml:"addr"`
	Log    string            `json:"log" yaml:"log"`
	Routes map[string]string `json:"routes" yaml:"routes"`
}

var Cfg *cfg = &cfg{}

func LoadCfg() (bool, error) {
	Cfg = &cfg{}
	bs, err := os.ReadFile("config.yaml")
	if err != nil {
		return false, err
	}
	err = yaml.Unmarshal(bs, Cfg)
	return true, err
}

func SaveCfg() error {
	bs, _ := yaml.Marshal(Cfg)
	return os.WriteFile("config.yaml", bs, 0600)
}

var Wd string
var UserHome string

func Init() error {
	var err error

	usr, err := user.Current()
	if err != nil {
		return err
	}
	UserHome = usr.HomeDir

	Wd = path.Join(usr.HomeDir, ".devserver")
	os.MkdirAll(Wd, os.ModePerm)
	os.Chdir(Wd)

	found, err := LoadCfg()
	if !found {
		Cfg.Addr = ":8443"
		Cfg.Log = "devserver.log"
		Cfg.Routes = map[string]string{
			"app.dev.local": "static://~/app",
			"api.dev.local": "http://localhost:8081",
			"raw.dev.local": "raw://~DS/raw",
		}
		SaveCfg()

	} else if err != nil {
		return err
	}

	if Cfg.Log != "-" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   Cfg.Log,
			MaxSize:    25, // megabytes
			MaxBackups: 10,
			MaxAge:     28,    //days
			Compress:   false, // disabled by default
		})
	}
	return err
}
