package goserver

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	Nats string `json:"nats"`
}

func ReadConfig(filename string) (*Config, error) {
	natsUrl := os.Getenv("GOSERVER_NATS_URL")
	if natsUrl != "" {
		logrus.WithField("natsUrl", natsUrl).Info("goserver")
		return &Config{Nats: natsUrl}, nil
	}
	if filename == "" {
		return &Config{Nats: nats.DefaultURL}, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	p := &Config{}
	err = json.NewDecoder(file).Decode(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func IsExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
