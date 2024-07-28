package main

import (
	"errors"
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
)

func getBuilds(path string) (builds []string, err error) {
	builds = []string{"ROOT"}
	err = filepath.WalkDir(path, func(dirPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && dirPath != path && d.Name() != "ROOT" {
			builds = append(builds, d.Name())
		}
		return err
	})
	return
}

func saveChanges(src string, dest string) (err error) {
	// Backup previous save
	// TODO: Remove unnecessary blocking
	err = copyFileContents(dest+"/"+MainFileName, dest+"/"+ErsmFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(dest+"/"+MainBackupFileName, dest+"/"+ErsmBackupFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(dest+"/"+MainSteamFileName, dest+"/"+ErsmSteamFileName)
	if err != nil {
		log.Println(err)
		return
	}
	// Moving files
	err = copyFileContents(src+"/"+MainFileName, dest+"/"+MainFileName)
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
	return
}

func createROOT(gameSavePath string, userSavePath string) (err error) {
	_, err = os.Stat(userSavePath + "\\ROOT")
	if err == nil {
		return
	}
	err = os.MkdirAll(userSavePath+"\\ROOT", 0777)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(gameSavePath+"\\"+MainFileName, userSavePath+"\\ROOT\\"+MainFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(gameSavePath+"\\"+MainBackupFileName, userSavePath+"\\ROOT\\"+MainBackupFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(gameSavePath+"\\"+MainSteamFileName, userSavePath+"\\ROOT\\"+MainSteamFileName)
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

	if currentInfo.Size() != backupInfo.Size() {
		return errors.New("invalid backup size")
	}

	err = os.Rename(savePath+"\\"+MainFileName, savePath+"\\temp.sl2")
	if err != nil {
		return
	}
	err = os.Rename(savePath+"\\"+MainBackupFileName, savePath+"\\temp.sl2.bak")
	if err != nil {
		return
	}
	err = os.Rename(savePath+"\\"+MainSteamFileName, savePath+"\\temp.vdf")
	if err != nil {
		return
	}

	err = os.Rename(savePath+"\\"+ErsmFileName, savePath+"\\"+MainFileName)
	if err != nil {
		return
	}
	err = os.Rename(savePath+"\\"+ErsmBackupFileName, savePath+"\\"+MainBackupFileName)
	if err != nil {
		return
	}
	err = os.Rename(savePath+"\\"+ErsmSteamFileName, savePath+"\\"+MainSteamFileName)
	if err != nil {
		return
	}

	err = os.Rename(savePath+"\\"+MainFileName, savePath+"\\"+ErsmFileName)
	if err != nil {
		return
	}
	err = os.Rename(savePath+"\\"+MainBackupFileName, savePath+"\\"+ErsmBackupFileName)
	if err != nil {
		return
	}
	err = os.Rename(savePath+"\\"+MainSteamFileName, savePath+"\\"+ErsmSteamFileName)
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
		return
	}
	err = out.Sync()
	return
}

func addBuild(userSavePath string, buildToBranchFrom string, newBuildName string) (err error) {
	err = os.Mkdir(userSavePath+"\\"+newBuildName, 0777)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(userSavePath+"\\"+buildToBranchFrom+"\\"+MainFileName, userSavePath+"\\"+newBuildName+"\\"+MainFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(userSavePath+"\\"+buildToBranchFrom+"\\"+MainBackupFileName, userSavePath+"\\"+newBuildName+"\\"+MainBackupFileName)
	if err != nil {
		log.Println(err)
		return
	}
	err = copyFileContents(userSavePath+"\\"+buildToBranchFrom+"\\"+MainSteamFileName, userSavePath+"\\"+newBuildName+"\\"+MainSteamFileName)
	return
}

func deleteBuild(userSavePath string, buildToDelete string) (err error) {
	err = os.RemoveAll(userSavePath + "\\" + buildToDelete)
	return
}
