package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"
	"image/color"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	a := app.New()
	w := a.NewWindow("Elden Ring Save Manager")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	gameSavePath := os.Getenv("GAME_SAVE_PATH")
	savesPath := os.Getenv("SAVE_FILES")
	data, err := os.ReadFile(savesPath + "\\ersm.cfg")
	if err != nil {
		log.Fatal(err)
	}
	currentBuild := string(data)

	// TODO: Handle error, likely dir not found
	builds, _ := getBuilds(savesPath)
	selectedBuild := ""
	var rollbackBtn *widget.Button
	hello := widget.NewLabel(currentBuild)
	buildSelector := widget.NewSelect(builds, func(value string) {
		log.Println("Select set to", value)
		selectedBuild = value
		rollbackBtn.SetText("Rollback \"" + value + "\" to Previous Save")
	})

	addBtn := widget.NewButton("Add New Build", func() {

	})
	loadBtn := widget.NewButton("Load", func() {
		hello.SetText(selectedBuild)
		// TODO: Handle case of current == selected
		err = saveChanges(gameSavePath, savesPath+"\\"+currentBuild)
		if err != nil {
			log.Fatal(err)
		}
		currentBuild = selectedBuild
		err = updateCurrentBuild(savesPath, currentBuild)
		if err != nil {
			log.Fatal(err)
		}
		err = loadFiles(savesPath+"\\"+currentBuild, gameSavePath)
		if err != nil {
			log.Fatal(err)
		}

	})
	rollbackBtn = widget.NewButton("Rollback to Previous Save", func() {
		prefix := savesPath + "\\" + selectedBuild
		backupInfo, err := os.Stat(prefix + "\\ERSM_backup.sl2")
		if err != nil {
			log.Println(err)
			return
		}
		currentInfo, err := os.Stat(prefix + "\\ER0000.sl2")
		if err != nil {
			log.Println(err)
			return
		}
		if backupInfo.Size() != currentInfo.Size() {
			var popup *widget.PopUp
			popup = widget.NewModalPopUp(container.NewVBox(
				canvas.NewText("Back up file is invalid. Unable to roll back to previous save.", color.White),
				widget.NewButton("Close", func() {
					popup.Hide()
				}),
			), w.Canvas())
			popup.Show()
			return
		}
		// TODO: Handle errors
		_ = copyFileContents(prefix+"\\ERSM_backup.sl2", prefix+"\\ER0000.sl2")
		_ = copyFileContents(prefix+"\\ERSM_backup.sl2.bak", prefix+"\\ER0000.sl2.bak")
		_ = copyFileContents(prefix+"\\ERSM_steam.vdf", prefix+"\\steam_autocloud.vdf")

	})

	mainCanvas := container.NewVBox(
		addBtn,
		buildSelector,
		hello,
		loadBtn,
		rollbackBtn,
	)

	buildSelector.SetSelected("ROOT")
	w.SetContent(mainCanvas)
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}

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
