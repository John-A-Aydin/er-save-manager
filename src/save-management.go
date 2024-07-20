package main

import (
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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
	err = copyFileContents(dest+"\\ER0000.sl2", dest+"\\ERSM_backup.sl2")
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(dest+"\\ER0000.sl2.bak", dest+"\\ERSM_backup.sl2.bak")
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(dest+"\\steam_autocloud.vdf", dest+"\\ERSM_steam.vdf")
	if err != nil {
		log.Println(err)
		return
	}
	// Moving files
	err = copyFileContents(src+"\\ER0000.sl2", dest+"\\ER0000.sl2")
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\ER0000.sl2.bak", dest+"\\ER0000.sl2.bak")
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\steam_autocloud.vdf", dest+"\\steam_autocloud.vdf")
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func loadFiles(src string, dest string) (err error) {
	err = copyFileContents(src+"\\ER0000.sl2", dest+"\\ER0000.sl2")
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\ER0000.sl2.bak", dest+"\\ER0000.sl2.bak")
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(src+"\\steam_autocloud.vdf", dest+"\\steam_autocloud.vdf")
	if err != nil {
		log.Println(err)
		return
	}
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

func updateCurrentBuild(savesPath string, value string) (err error) {
	f, err := os.Create(savesPath + "\\current.ersm")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = f.WriteString(value)
	if err != nil {
		log.Println(err)
		return
	}
	return f.Close()
}

func writeConfig(cfg Config) (err error) {
	dat, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}
	err = os.WriteFile("config.yaml", dat, 0644)
	if err != nil {
		return
	}
	return
}

func readConfig() (cfg Config, err error) {
	dat, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println(err)
		log.Println("adklfjalkdsf;lkasdfjklas")
		return Config{}, err
	}
	err = yaml.Unmarshal([]byte(dat), &cfg)
	if err != nil {
		log.Println(err)
	}
	log.Println(cfg.CurrentBuild)
	return cfg, err
}
