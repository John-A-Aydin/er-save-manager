package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
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

	var rollbackBtn *widget.Button
	var loadedBuildIndicator *widget.Label
	var buildSelector *widget.Select
	var loadBtn *widget.Button
	var mainContainer *fyne.Container
	var toolbar *widget.Toolbar

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
			{Text: "User Save Path:", Widget: playerSavePathEntry},
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

	// TODO: Add branching from different saves
	buildNameInput := widget.NewEntry()
	addForm := widget.Form{
		Items: []*widget.FormItem{
			{Text: "Build Name:", Widget: buildNameInput},
			//{Widget: setAsActive},
		},
		OnSubmit: func() {
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
			err = addBuild(cfg.UserSavePath, newBuildName)
			if err != nil {
				log.Println(err)
				w.SetContent(mainContainer)
				return
			}
			builds = append(builds, newBuildName)
			buildSelector = widget.NewSelect(builds, func(value string) {
				log.Println("Select set to", value)
				rollbackBtn.SetText("Rollback \"" + value + "\" to Previous Save")
			})
			buildSelector.SetSelected(newBuildName)
			mainContainer = container.NewVBox(
				toolbar,
				buildSelector,
				loadedBuildIndicator,
				loadBtn,
				rollbackBtn,
			)
			w.SetContent(mainContainer)
			log.Println(newBuildName)
		},
		OnCancel: func() {
			w.SetContent(mainContainer)
		},
	}
	addForm.SubmitText = "Create"

	loadedBuildIndicator = widget.NewLabel("Currently Loaded: " + cfg.CurrentBuild)
	buildSelector = widget.NewSelect(builds, func(value string) {
		log.Println("Select set to", value)
		rollbackBtn.SetText("Rollback \"" + value + "\" to Previous Save")
	})

	loadBtn = widget.NewButton("Load", func() {
		// TODO: Handle case of current == selected
		if cfg.CurrentBuild == buildSelector.Selected {
			return
		}
		cfg.CurrentBuild = buildSelector.Selected
		loadedBuildIndicator.SetText("Currently Loaded: " + buildSelector.Selected)
		currentBuildSelector.SetSelected(buildSelector.Selected)
		err = saveChanges(cfg.GameSavePath, cfg.UserSavePath+"\\"+cfg.CurrentBuild)
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
		err = rollBackSave(cfg.UserSavePath + "\\" + cfg.CurrentBuild)
		if err != nil && err.Error() == "mismatched file size" {
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

	})

	toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			w.SetContent(&settingsForm)
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			w.SetContent(&addForm)
		}))

	mainContainer = container.NewVBox(
		toolbar,
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
