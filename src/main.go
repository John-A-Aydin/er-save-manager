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
	"strings"
)

func main() {
	a := app.New()
	icon, err := fyne.LoadResourceFromPath("assets\\icon.png")
	a.SetIcon(icon)
	w := a.NewWindow("Elden Ring Save Manager")

	if !findConfig() {
		err = createConfig()
		if err != nil {
			panic(err)
		}
	}
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}

	// TODO: Handle error, likely dir not found
	builds, _ := getBuilds(cfg.UserSavePath)

	var mainContainer *fyne.Container
	var toolbar *widget.Toolbar
	var loadedBuildIndicator *widget.Label
	var loadAndSaveBtn *widget.Button
	var loadWithoutSaveBtn *widget.Button
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
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			_ = saveChanges(cfg.GameSavePath, cfg.UserSavePath+"\\"+cfg.CurrentBuild)
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			w.SetContent(addForm)
		}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			w.SetContent(deleteForm)
		}),
	)

	loadedBuildIndicator = widget.NewLabel("Currently Loaded: " + cfg.CurrentBuild)

	loadAndSaveBtn = widget.NewButton("Load + Save", func() {
		err = saveChanges(cfg.GameSavePath, cfg.UserSavePath+"\\"+cfg.CurrentBuild)
		if err != nil {
			log.Fatal(err)
		}
		if cfg.CurrentBuild == mainBuildSelector.Selected {
			return
		}
		cfg.CurrentBuild = mainBuildSelector.Selected
		err = writeConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		loadedBuildIndicator.SetText("Currently Loaded: " + cfg.CurrentBuild)
		settingsBuildSelector.SetSelected(cfg.CurrentBuild)
		err = loadFiles(cfg.UserSavePath+"\\"+cfg.CurrentBuild, cfg.GameSavePath)
		if err != nil {
			log.Fatal(err)
		}
	})

	loadWithoutSaveBtn = widget.NewButton("Load w/o Saving", func() {
		cfg.CurrentBuild = mainBuildSelector.Selected
		err = writeConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		loadedBuildIndicator.SetText("Currently Loaded: " + cfg.CurrentBuild)
		settingsBuildSelector.SetSelected(cfg.CurrentBuild)
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
		loadAndSaveBtn,
		loadWithoutSaveBtn,
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
			err = writeConfig(cfg)
			if err != nil {
				log.Fatal(err)
			}
			w.SetContent(mainContainer)
		})))
	if cfg.UserSavePath == "" {
		w.SetContent(initializeUserSavePathForm)
	} else {
		w.SetContent(mainContainer)
	}

	w.Resize(fyne.NewSize(400, 300))

	w.ShowAndRun()
}
