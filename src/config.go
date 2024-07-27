package main

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
)

const (
	ConfigFileName   = "config.yaml"
	ConfigFolderName = "EldenRingSaveManager"
)

type Config struct {
	GameSavePath string `yaml:"game-save-path"`
	UserSavePath string `yaml:"user-save-path"`
	CurrentBuild string `yaml:"current-build"`
}

func writeConfig(cfg Config) (err error) {
	dat, err := yaml.Marshal(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		log.Println(err)
		return
	}
	err = os.WriteFile(cfgDir+"\\"+ConfigFolderName+"\\"+ConfigFileName, dat, 0644)
	return
}

func readConfig() (cfg Config, err error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}
	dat, err := os.ReadFile(cfgDir + "\\" + ConfigFolderName + "\\" + ConfigFileName)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal([]byte(dat), &cfg)
	if err != nil {
		return Config{}, err
	}
	return
}

func findConfig() bool {
	dir, err := os.UserConfigDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(dir + "\\" + ConfigFolderName)
	if err != nil {
		return false
	}
	return true
}

func createConfig() (err error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Println(err)
		return
	}
	ersmDir := dir + "\\" + ConfigFolderName
	err = os.Mkdir(ersmDir, 0777)
	if err != nil {
		log.Println(err)
		return
	}
	gameSaveDir := dir + "\\EldenRing"
	entries, err := os.ReadDir(gameSaveDir)
	if err != nil {
		log.Println(err)
		return
	}
	for _, entry := range entries {
		_, err = strconv.Atoi(entry.Name())
		if entry.IsDir() && err == nil {
			gameSaveDir = gameSaveDir + "\\" + entry.Name()
			break
		}
	}
	cfg := Config{
		UserSavePath: "",
		GameSavePath: gameSaveDir,
		CurrentBuild: "ROOT",
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	err = os.WriteFile(ersmDir+"\\"+ConfigFileName, data, 0644)
	return
}
