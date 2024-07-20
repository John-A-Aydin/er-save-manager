package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"os"
	"strings"
)

type Config struct {
	GameSavePath string `yaml:"game-save-path"`
	UserSavePath string `yaml:"user-save-path"`
	CurrentBuild string `yaml:"current-build"`
}

func main() {
	a := app.New()
	icon, err := fyne.LoadResourceFromPath("assets\\icon.png")
	a.SetIcon(icon)
	w := a.NewWindow("Elden Ring Save Manager")

	cfg, err := readConfig()
	if err != nil {
		log.Println("Unable to read config file")
		log.Fatal(err)
	}

	//// TODO: Handle error, likely dir not found
	builds, _ := getBuilds(cfg.UserSavePath)
	selectedBuild := ""

	var settingsBtn *widget.Button
	var rollbackBtn *widget.Button
	var loadedBuildIndicator *widget.Label
	var addBtn *widget.Button
	var buildSelector *widget.Select
	var loadBtn *widget.Button
	var mainContainer *fyne.Container

	gameSavePathEntry := widget.NewEntry()
	gameSavePathEntry.SetPlaceHolder(cfg.GameSavePath)
	playerSavePathEntry := widget.NewEntry()
	playerSavePathEntry.SetPlaceHolder(cfg.UserSavePath)
	currentBuildSelector := widget.NewSelect(builds, func(value string) {
		log.Println("Select set to", value)
	})
	currentBuildSelector.SetSelected(cfg.CurrentBuild)
	settingsForm := widget.Form{
		Items: []*widget.FormItem{
			{Text: "Game Save Path:", Widget: gameSavePathEntry},
			{Text: "Your Save Path:", Widget: playerSavePathEntry},
			{Text: "Currently Loaded Build:", Widget: currentBuildSelector},
		},
		OnSubmit: func() {
			log.Println(gameSavePathEntry.Text)
			cfg.CurrentBuild = currentBuildSelector.Selected
			loadedBuildIndicator.SetText("Currently Loaded: " + cfg.CurrentBuild)
			buildSelector.SetSelected(cfg.CurrentBuild)
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

	loadedBuildIndicator = widget.NewLabel("Currently Loaded: " + cfg.CurrentBuild)
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
				err := os.Mkdir(cfg.UserSavePath+"\\"+newBuildName, 0777)
				if err != nil {
					return
				}
				builds = append(builds, newBuildName)
				err = copyFileContents(cfg.UserSavePath+"\\ROOT\\ER0000.sl2", cfg.UserSavePath+"\\"+newBuildName+"\\ER0000.sl2")
				if err != nil {
					log.Fatal(err)
				}
				err = copyFileContents(cfg.UserSavePath+"\\ROOT\\ER0000.sl2.bak", cfg.UserSavePath+"\\"+newBuildName+"\\ER0000.sl2.bak")
				if err != nil {
					log.Fatal(err)
				}
				err = copyFileContents(cfg.UserSavePath+"\\ROOT\\steam_autocloud.vdf", cfg.UserSavePath+"\\"+newBuildName+"\\steam_autocloud.vdf")
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
					loadedBuildIndicator,
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
		if cfg.CurrentBuild == selectedBuild {
			return
		}
		loadedBuildIndicator.SetText("Currently Loaded: " + selectedBuild)
		currentBuildSelector.SetSelected(selectedBuild)
		err = saveChanges(cfg.GameSavePath, cfg.UserSavePath+"\\"+cfg.CurrentBuild)
		if err != nil {
			log.Fatal(err)
		}
		cfg.CurrentBuild = selectedBuild
		err = writeConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		err = writeConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		err = loadFiles(cfg.UserSavePath+"\\"+cfg.CurrentBuild, cfg.GameSavePath)
		if err != nil {
			log.Fatal(err)
		}

	})
	rollbackBtn = widget.NewButton("Rollback to Previous Save", func() {
		prefix := cfg.UserSavePath + "\\" + selectedBuild
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
		loadedBuildIndicator,
		loadBtn,
		rollbackBtn,
	)

	buildSelector.SetSelected(cfg.CurrentBuild)
	w.SetContent(mainContainer)
	w.Resize(fyne.NewSize(400, 300))

	w.ShowAndRun()
}
