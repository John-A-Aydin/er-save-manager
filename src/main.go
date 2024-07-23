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
	icon, err := fyne.LoadResourceFromPath("assets/icon.png")
	a.SetIcon(icon)
	w := a.NewWindow("Elden Ring Save Manager")

	cfg, err := readConfig()
	if err != nil {
		log.Println("Unable to read config file")
		log.Fatal(err)
	}

	//// TODO: Handle error, likely dir not found
	builds, _ := getBuilds(cfg.UserSavePath)

	var mainContainer *fyne.Container
	var toolbar *widget.Toolbar
	var loadedBuildIndicator *widget.Label
	var buildSelector *widget.Select
	var loadBtn *widget.Button
	var rollbackBtn *widget.Button
	var addForm *widget.Form
	var deleteForm *widget.Form
	var settingsForm *widget.Form

	var createSettingsForm func() *widget.Form
	var createAddForm func() *widget.Form
	var createDeleteForm func() *widget.Form

	buildSelector = widget.NewSelect(builds, func(value string) {
		log.Println("Select set to", value)
		rollbackBtn.SetText("Rollback \"" + value + "\" to Previous Save")
	})

	gameSavePathEntry := widget.NewEntry()
	gameSavePathEntry.SetPlaceHolder(cfg.GameSavePath)
	playerSavePathEntry := widget.NewEntry()
	playerSavePathEntry.SetPlaceHolder(cfg.UserSavePath)
	createSettingsForm = func() *widget.Form {
		return &widget.Form{
			Items: []*widget.FormItem{
				{Text: "Game Save Path:", Widget: gameSavePathEntry},
				{Text: "User Save Path:", Widget: playerSavePathEntry},
				{Text: "Currently Loaded Build:", Widget: buildSelector},
			},
			OnSubmit: func() {
				log.Println(gameSavePathEntry.Text)
				cfg.CurrentBuild = buildSelector.Selected
				loadedBuildIndicator.SetText("Currently Loaded: " + cfg.CurrentBuild)
				buildSelector.SetSelected(cfg.CurrentBuild)
				gameSavePathEntry.SetText("")
				playerSavePathEntry.SetText("")
				w.SetContent(mainContainer)
			},
			OnCancel: func() {
				gameSavePathEntry.SetText("")
				playerSavePathEntry.SetText("")
				w.SetContent(mainContainer)
			},
		}
	}

	settingsForm = createSettingsForm()
	settingsForm.SubmitText = "Save"

	// TODO: Add branching from different saves
	buildNameEntry := widget.NewEntry()
	createAddForm = func() *widget.Form {
		return &widget.Form{
			Items: []*widget.FormItem{
				{Text: "Build Name:", Widget: buildNameEntry},
			},
			OnSubmit: func() {
				newBuildName := strings.Trim(buildNameEntry.Text, " ")
				// TODO: Handle user notification in error cases
				if newBuildName == "" || strings.Contains(newBuildName, "/") {
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
				})
				buildSelector.SetSelected(newBuildName)
				settingsForm = createSettingsForm()
				deleteForm = createDeleteForm()
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
				mainContainer = container.NewVBox(
					toolbar,
					buildSelector,
					loadedBuildIndicator,
					loadBtn,
					rollbackBtn,
				)
				buildNameEntry.SetText("")
				w.SetContent(mainContainer)
			},
			OnCancel: func() {
				buildNameEntry.SetText("")
				w.SetContent(mainContainer)
			},
		}
	}
	addForm = createAddForm()

	addForm.SubmitText = "Create"
	deleteFormTextConfirmation := widget.NewEntry()
	deleteFormTextConfirmation.SetPlaceHolder("Confirm by typing \"delete\"")
	createDeleteForm = func() *widget.Form {
		return &widget.Form{
			Items: []*widget.FormItem{
				{Text: "", Widget: buildSelector},
				{Text: "", Widget: deleteFormTextConfirmation},
			},
			OnSubmit: func() { // TODO: Actually delete the files
				if deleteFormTextConfirmation.Text != "delete" {
					return
				}
				buildToDelete := buildSelector.Selected
				newSelectedBuild := buildSelector.Selected
				// TODO: Notify that the root build cannot be deleted
				if buildToDelete == "ROOT" {
					return
				}
				if buildToDelete == cfg.CurrentBuild {
					err = loadFiles(cfg.UserSavePath+"/ROOT", cfg.GameSavePath)
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
				deleteFormTextConfirmation.SetText("")
				buildSelector = widget.NewSelect(builds, func(value string) {
					log.Println("Select set to", value)
				})
				buildSelector.SetSelected(newSelectedBuild)
				settingsForm = createSettingsForm()
				deleteForm = createDeleteForm()
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
				mainContainer = container.NewVBox(
					toolbar,
					buildSelector,
					loadedBuildIndicator,
					loadBtn,
					rollbackBtn,
				)
				w.SetContent(mainContainer)
			},
			OnCancel: func() {
				deleteFormTextConfirmation.SetText("")
				w.SetContent(mainContainer)
			},
		}
	}
	deleteForm = createDeleteForm()
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
		if cfg.CurrentBuild == buildSelector.Selected {
			return
		}
		cfg.CurrentBuild = buildSelector.Selected
		loadedBuildIndicator.SetText("Currently Loaded: " + buildSelector.Selected)
		err = saveChanges(cfg.GameSavePath, cfg.UserSavePath+"/"+cfg.CurrentBuild)
		if err != nil {
			log.Fatal(err)
		}
		err = writeConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		err = loadFiles(cfg.UserSavePath+"/"+cfg.CurrentBuild, cfg.GameSavePath)
		if err != nil {
			log.Fatal(err)
		}
	})
	rollbackBtn = widget.NewButton("Rollback to Previous Save", func() {
		err = rollBackSave(cfg.UserSavePath + "/" + cfg.CurrentBuild)
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
