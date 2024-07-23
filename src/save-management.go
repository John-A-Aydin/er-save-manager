package main

import (
	"errors"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

const (
	MainFileName       = "ER0000.sl2"
	MainBackupFileName = "ER0000.sl2.bak"
	MainSteamFileName  = "steam_autocloud.vdf"
	ErsmFileName       = "ERSM_backup.sl2"
	ErsmBackupFileName = "ERSM_backup.sl2.bak"
	ErsmSteamFileName  = "ERSM_steam.vdf"
	ConfigFileName     = "config.yaml"
)

func getBuilds(path string) ([]string, error) {
	builds := []string{"ROOT"}
	err := filepath.WalkDir(path, func(dirPath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if d.IsDir() && dirPath != path && d.Name() != "ROOT" {
			builds = append(builds, d.Name())
		}
		return err
	})
	return builds, err
}

func saveChanges(src string, dest string) (err error) {
	// Backup previous save
	// TODO: Remove unnecessary blocking
	err = copyFileContents(dest+"\\"+MainFileName, dest+"\\"+ErsmFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(dest+"\\"+MainBackupFileName, dest+"\\"+ErsmBackupFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(dest+"\\"+MainSteamFileName, dest+"\\"+ErsmSteamFileName)
	if err != nil {
		log.Println(err)
		return
	}
	// Moving files
	err = copyFileContents(src+"\\"+MainFileName, dest+"\\"+MainFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\"+MainBackupFileName, dest+"\\"+MainBackupFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\"+MainSteamFileName, dest+"\\"+MainSteamFileName)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func loadFiles(src string, dest string) (err error) {
	err = copyFileContents(src+"\\"+MainFileName, dest+"\\"+MainFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\"+MainBackupFileName, dest+"\\"+MainBackupFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\"+MainSteamFileName, dest+"\\"+MainSteamFileName)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func rollBackSave(savePath string) (err error) {
	backupInfo, err := os.Stat(savePath + "\\" + ErsmFileName)
	if err != nil {
		log.Println(err)
		return
	}
	currentInfo, err := os.Stat(savePath + "\\" + MainFileName)
	if err != nil {
		log.Println(err)
		return
	}
	if backupInfo.Size() != currentInfo.Size() {
		return errors.New("mismatched file size")
	}
	err = copyFileContents(savePath+"\\"+ErsmFileName, savePath+"\\"+MainFileName)
	if err != nil {
		return
	}
	err = copyFileContents(savePath+"\\"+ErsmBackupFileName, savePath+"\\"+MainBackupFileName)
	if err != nil {
		return
	}
	err = copyFileContents(savePath+"\\"+ErsmSteamFileName, savePath+"\\"+MainSteamFileName)
	return
}

func copyFileContents(src string, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		log.Println(err)
		return
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		log.Println(err)
		return
	}
	err = out.Sync()
	return
}

func writeConfig(cfg Config) (err error) {
	dat, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}
	err = os.WriteFile(ConfigFileName, dat, 0644)
	if err != nil {
		return
	}
	return
}

func readConfig() (cfg Config, err error) {
	dat, err := os.ReadFile(ConfigFileName)
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	err = yaml.Unmarshal([]byte(dat), &cfg)
	if err != nil {
		log.Println(err)
	}
	log.Println(cfg.CurrentBuild)
	return cfg, err
}

func addBuild(userSavePath string, newBuildName string) (err error) {
	err = os.Mkdir(userSavePath+"\\"+newBuildName, 0777)
	if err != nil {
		return
	}
	err = copyFileContents(userSavePath+"\\ROOT\\"+MainFileName, userSavePath+"\\"+newBuildName+"\\"+MainFileName)
	if err != nil {
		return
	}
	err = copyFileContents(userSavePath+"\\ROOT\\"+MainBackupFileName, userSavePath+"\\"+newBuildName+"\\"+MainBackupFileName)
	if err != nil {
		return
	}
	err = copyFileContents(userSavePath+"\\ROOT\\"+MainSteamFileName, userSavePath+"\\"+newBuildName+"\\"+MainSteamFileName)
	return
}

func deleteBuild(userSavePath string, buildToDelete string) (err error) {
	err = os.RemoveAll(userSavePath + "\\" + buildToDelete)
	return
}
