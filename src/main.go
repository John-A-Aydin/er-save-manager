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
	"os"
	"strconv"
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

	// TODO: Handle error, likely dir not found
	builds, _ := getBuilds(cfg.UserSavePath)

	var mainContainer *fyne.Container
	var toolbar *widget.Toolbar
	var loadedBuildIndicator *widget.Label
	var loadBtn *widget.Button
	var rollbackBtn *widget.Button
	var addForm *widget.Form
	var deleteForm *widget.Form
	var settingsForm *widget.Form

	mainBuildSelector := widget.NewSelect(builds, func(value string) {})
	settingsBuildSelector := widget.NewSelect(builds, func(value string) {})
	deleteBuildSelector := widget.NewSelect(builds[1:], func(value string) {})
	addBuildSelector := widget.NewSelect(builds, func(value string) {})

	mainBuildSelector.SetSelected(cfg.CurrentBuild)
	settingsBuildSelector.SetSelected(cfg.CurrentBuild)
	addBuildSelector.SetSelected("ROOT")

	gameSavePathEntry := widget.NewEntry()
	gameSavePathEntry.SetPlaceHolder(cfg.GameSavePath)
	userSavePathEntry := widget.NewEntry()
	userSavePathEntry.SetPlaceHolder(cfg.UserSavePath)
	settingsForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Game Save Path:", Widget: gameSavePathEntry},
			{Text: "User Save Path:", Widget: userSavePathEntry},
			{Text: "Currently Loaded Build:", Widget: settingsBuildSelector},
		},
		OnSubmit: func() { // TODO: Notify user if the path is invalid
			if gameSavePathEntry.Text != "" {
				_, err = os.Stat(gameSavePathEntry.Text)
				if err == nil {
					cfg.GameSavePath = gameSavePathEntry.Text
					gameSavePathEntry.SetPlaceHolder(cfg.GameSavePath)
				}
			}
			if userSavePathEntry.Text != "" {
				_, err = os.Stat(userSavePathEntry.Text)
				if err == nil {
					cfg.UserSavePath = userSavePathEntry.Text
					userSavePathEntry.SetPlaceHolder(cfg.UserSavePath)
					builds, _ = getBuilds(cfg.UserSavePath)
					mainBuildSelector.SetOptions(builds)
					settingsBuildSelector.SetOptions(builds)
					deleteBuildSelector.SetOptions(builds[1:])
					addBuildSelector.SetOptions(builds)

					mainBuildSelector.SetSelected(cfg.CurrentBuild)
					settingsBuildSelector.SetSelected(cfg.CurrentBuild)
					addBuildSelector.SetSelected("ROOT")
				}
			}
			cfg.CurrentBuild = settingsBuildSelector.Selected
			err = writeConfig(cfg)
			if err != nil {
				log.Fatal(err)
			}
			loadedBuildIndicator.SetText("Currently Loaded: " + cfg.CurrentBuild)
			mainBuildSelector.SetSelected(cfg.CurrentBuild)
			gameSavePathEntry.SetText("")
			userSavePathEntry.SetText("")
			w.SetContent(mainContainer)
		},
		OnCancel: func() {
			gameSavePathEntry.SetText("")
			userSavePathEntry.SetText("")
			settingsBuildSelector.SetSelected(cfg.CurrentBuild)
			w.SetContent(mainContainer)
		},
	}
	settingsForm.SubmitText = "Save"

	buildNameEntry := widget.NewEntry()
	addForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Build Name:", Widget: buildNameEntry},
			{Text: "Build to Branch From:", Widget: addBuildSelector},
		},
		OnSubmit: func() {
			newBuildName := strings.Trim(buildNameEntry.Text, " ")
			// TODO: Handle user notification in error cases
			if newBuildName == "" || strings.Contains(newBuildName, "\\") || strings.Contains(newBuildName, "\\") {
				return
			}
			for _, build := range builds {
				if build == newBuildName {
					return
				}
			}
			err = addBuild(cfg.UserSavePath, addBuildSelector.Selected, newBuildName)
			if err != nil {
				log.Println(err)
				w.SetContent(mainContainer)
				return
			}
			builds = append(builds, newBuildName)
			mainBuildSelector.SetOptions(builds)
			settingsBuildSelector.SetOptions(builds)
			deleteBuildSelector.SetOptions(builds[1:])
			addBuildSelector.SetOptions(builds)
			addBuildSelector.SetSelected("ROOT")
			buildNameEntry.SetText("")
			w.SetContent(mainContainer)
		},
		OnCancel: func() {
			buildNameEntry.SetText("")
			w.SetContent(mainContainer)
		},
	}

	addForm.SubmitText = "Create"
	deleteFormTextConfirmation := widget.NewEntry()
	deleteFormTextConfirmation.SetPlaceHolder("Confirm by typing \"delete\"")
	deleteForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "", Widget: deleteBuildSelector},
			{Text: "", Widget: deleteFormTextConfirmation},
		},
		OnSubmit: func() { // TODO: Actually delete the files
			if deleteFormTextConfirmation.Text != "delete" {
				return
			}
			buildToDelete := deleteBuildSelector.Selected
			newSelectedBuild := mainBuildSelector.Selected
			// TODO: Notify that the root build cannot be deleted
			if buildToDelete == "ROOT" || buildToDelete == "" {
				return
			}
			if buildToDelete == cfg.CurrentBuild {
				err = loadFiles(cfg.UserSavePath+"\\ROOT", cfg.GameSavePath)
				if err != nil {
					log.Fatal(err)
				}
				loadedBuildIndicator.SetText("Currently Loaded: ROOT")
				cfg.CurrentBuild = "ROOT"
				err = writeConfig(cfg)
				if err != nil {
					log.Fatal(err)
				}
				newSelectedBuild = "ROOT"
			}
			idxToDelete := -1
			for idx, build := range builds {
				if build == buildToDelete {
					idxToDelete = idx
				}
			}
			if idxToDelete == -1 {
				panic("BuildToDelete was not found")
			}
			builds = append(builds[:idxToDelete], builds[idxToDelete+1:]...)
			err = deleteBuild(cfg.UserSavePath, buildToDelete)
			if err != nil {
				log.Fatal(err)
			}
			if newSelectedBuild == buildToDelete {
				newSelectedBuild = "ROOT"
			}
			mainBuildSelector.SetOptions(builds)
			mainBuildSelector.SetSelected(newSelectedBuild)
			settingsBuildSelector.SetOptions(builds)
			settingsBuildSelector.SetSelected(cfg.CurrentBuild)
			deleteBuildSelector.SetOptions(builds[1:])
			deleteBuildSelector.ClearSelected()
			addBuildSelector.SetOptions(builds)

			deleteFormTextConfirmation.SetText("")
			w.SetContent(mainContainer)
		},
		OnCancel: func() {
			deleteFormTextConfirmation.SetText("")
			deleteBuildSelector.ClearSelected()
			w.SetContent(mainContainer)
		},
	}
	deleteForm.SubmitText = "Delete"

	toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			w.SetContent(settingsForm)
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			w.SetContent(addForm)
		}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			w.SetContent(deleteForm)
		}),
	)

	loadedBuildIndicator = widget.NewLabel("Currently Loaded: " + cfg.CurrentBuild)

	loadBtn = widget.NewButton("Load", func() {
		if cfg.CurrentBuild == mainBuildSelector.Selected {
			return
		}
		cfg.CurrentBuild = mainBuildSelector.Selected
		loadedBuildIndicator.SetText("Currently Loaded: " + mainBuildSelector.Selected)
		settingsBuildSelector.SetSelected(cfg.CurrentBuild)
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

	mainContainer = container.NewVBox(
		toolbar,
		mainBuildSelector,
		loadedBuildIndicator,
		loadBtn,
		rollbackBtn,
	)

	pathEntry := widget.NewEntry()
	initializeUserSavePathForm := container.NewVBox(container.NewVBox(
		widget.NewLabel("Create a folder to store your saves and paste the path below"),
		pathEntry,
		widget.NewButton("Save", func() {
			if pathEntry.Text != "" {
				_, err = os.Stat(pathEntry.Text)
				if err != nil {
					return
				}
				cfg.UserSavePath = pathEntry.Text
			}
			_ = writeConfig(cfg)
			builds, _ = getBuilds(cfg.UserSavePath)

			mainBuildSelector.SetOptions(builds)
			settingsBuildSelector.SetOptions(builds)
			deleteBuildSelector.SetOptions(builds[1:])
			addBuildSelector.SetOptions(builds)

			mainBuildSelector.SetSelected(cfg.CurrentBuild)
			settingsBuildSelector.SetSelected(cfg.CurrentBuild)
			addBuildSelector.SetSelected("ROOT")

			userSavePathEntry.SetPlaceHolder(cfg.UserSavePath)

			_, err = os.Stat(cfg.UserSavePath + "\\ROOT")
			if err != nil {
				_ = createROOT(cfg.GameSavePath, cfg.UserSavePath)
			}

			w.SetContent(mainContainer)
		})))
	if cfg.GameSavePath == "" {
		usrDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		saveDir := usrDir + "\\AppData\\Roaming\\EldenRing"
		entries, err := os.ReadDir(saveDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			// The target folder is a numeric SteamID
			_, err = strconv.Atoi(entry.Name())
			if entry.IsDir() && err == nil {
				saveDir = saveDir + "\\" + entry.Name()
				break
			}
		}
		cfg.GameSavePath = saveDir
		gameSavePathEntry.SetPlaceHolder(saveDir)
		_ = writeConfig(cfg)
	}
	if cfg.UserSavePath == "" {
		w.SetContent(initializeUserSavePathForm)
	} else {
		w.SetContent(mainContainer)
	}

	w.Resize(fyne.NewSize(400, 300))

	w.ShowAndRun()
}
