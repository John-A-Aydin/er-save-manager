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
	"strings"
)

func main() {
	a := app.New()
	icon, err := fyne.LoadResourceFromPath("assets\\icon.png")
	a.SetIcon(icon)
	w := a.NewWindow("Elden Ring Save Manager")
	err = godotenv.Load()
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

	var settingsBtn *widget.Button
	var rollbackBtn *widget.Button
	var hello *widget.Label
	var addBtn *widget.Button
	var buildSelector *widget.Select
	var loadBtn *widget.Button
	var mainContainer *fyne.Container
	//var currentBuildSelector *widget.Select

	gameSavePathEntry := widget.NewEntry()
	gameSavePathEntry.SetPlaceHolder(gameSavePath)
	playerSavePathEntry := widget.NewEntry()
	playerSavePathEntry.SetPlaceHolder(savesPath)
	settingsSelectedBuild := currentBuild
	currentBuildSelector := widget.NewSelect(builds, func(value string) {
		log.Println("Select set to", value)
		settingsSelectedBuild = value
	})
	currentBuildSelector.SetSelected(currentBuild)
	settingsForm := widget.Form{
		Items: []*widget.FormItem{
			{Text: "Game Save Path:", Widget: gameSavePathEntry},
			{Text: "Your Save Path:", Widget: playerSavePathEntry},
			{Text: "Currently Loaded Build:", Widget: currentBuildSelector},
		},
		OnSubmit: func() {
			log.Println(gameSavePathEntry.Text)
			currentBuildSelector.Selected = settingsSelectedBuild
			hello.SetText("Currently Loaded: " + settingsSelectedBuild)
			w.SetContent(mainContainer)
		},
		OnCancel: func() {
			w.SetContent(mainContainer)
		},
	}
	settingsForm.SubmitText = "Save"

	settingsBtn = widget.NewButton("Settings", func() {
		w.SetContent(&settingsForm)
	})

	hello = widget.NewLabel("Currently Loaded: " + currentBuild)
	buildSelector = widget.NewSelect(builds, func(value string) {
		log.Println("Select set to", value)
		selectedBuild = value
		rollbackBtn.SetText("Rollback \"" + value + "\" to Previous Save")
	})
	// TODO: Add feature to branch from different saves.
	addBtn = widget.NewButton("Add New Build", func() {
		var popup *widget.PopUp
		buildNameInput := widget.NewEntry()
		//buildNameInput.SetPlaceHolder("Enter Build Name...")

		//setAsActive := widget.NewRadioGroup([]string{"Load New Build"}, func(value string) {
		//	log.Println("Select set to", value)
		//})
		form := &widget.Form{
			Items: []*widget.FormItem{
				{Text: "Build Name:", Widget: buildNameInput},
				//{Widget: setAsActive},
			},
			OnSubmit: func() {
				defer popup.Hide()
				newBuildName := strings.Trim(buildNameInput.Text, " ")
				// TODO: Handle user notification in error cases
				if newBuildName == "" || strings.Contains(newBuildName, "\\") {
					return
				}
				for _, build := range builds {
					if build == newBuildName {
						return
					}
				}
				err := os.Mkdir(savesPath+"\\"+newBuildName, 0777)
				if err != nil {
					return
				}
				builds = append(builds, newBuildName)
				err = copyFileContents(savesPath+"\\ROOT\\ER0000.sl2", savesPath+"\\"+newBuildName+"\\ER0000.sl2")
				if err != nil {
					log.Fatal(err)
				}
				err = copyFileContents(savesPath+"\\ROOT\\ER0000.sl2.bak", savesPath+"\\"+newBuildName+"\\ER0000.sl2.bak")
				if err != nil {
					log.Fatal(err)
				}
				err = copyFileContents(savesPath+"\\ROOT\\steam_autocloud.vdf", savesPath+"\\"+newBuildName+"\\steam_autocloud.vdf")
				if err != nil {
					log.Fatal(err)
				}
				buildSelector = widget.NewSelect(builds, func(value string) {
					log.Println("Select set to", value)
					selectedBuild = value
					rollbackBtn.SetText("Rollback \"" + value + "\" to Previous Save")
				})
				buildSelector.SetSelected(newBuildName)
				selectedBuild = newBuildName
				mainContainer = container.NewVBox(
					addBtn,
					buildSelector,
					hello,
					loadBtn,
					rollbackBtn,
				)
				w.SetContent(mainContainer)
				log.Println(newBuildName)
			},
			OnCancel: func() {
				popup.Hide()
			},
		}
		form.SubmitText = "Create"

		popup = widget.NewModalPopUp(container.NewVBox(
			form,
		), w.Canvas())

		popup.Show()

	})
	loadBtn = widget.NewButton("Load", func() {
		// TODO: Handle case of current == selected
		if currentBuild == selectedBuild {
			return
		}
		hello.SetText("Currently Loaded: " + selectedBuild)
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

				canvas.NewText("Unable to rollback", color.White),
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

	mainContainer = container.NewVBox(
		settingsBtn,
		addBtn,
		buildSelector,
		hello,
		loadBtn,
		rollbackBtn,
	)

	buildSelector.SetSelected("ROOT")
	w.SetContent(mainContainer)
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
