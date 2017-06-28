package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

func (c *C) ReadConfig(configPath string) error {
	log.Debug("Reading configuration...")

	if _, err := toml.DecodeFile(configPath, c); err != nil {
		return err
	}

	return nil
}

func (c *C) WriteConfig(configPath string) error {
	log.Debug("Writing configuration...")

	if configPath == "" {
		return errors.New("No configuration path given!")
	}

	file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	encoder.Encode(c)

	return nil
}

func (c *C) SetDefaults() {
	log.Debug("Overwriting loaded configuration with defaults!")

	c.Pools = make([]*Pool, 0)
	c.Workers = make([]*Worker, 0)
}
