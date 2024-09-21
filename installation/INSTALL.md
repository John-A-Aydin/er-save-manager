# How to install

## Disclaimer
It is recommended that you make a separate copy of your root save file in a folder that will not be touched by this program.

It's possible to overwrite the contents of your root save file with another save file if you change which save is being tracked in the settings. There is a rollback button to revert the last change made to a save file, but this action is destructive and will delete any changes made.

## Download

Download the windows installer [here](https://downgit.github.io/#/home?url=https://github.com/John-A-Aydin/er-save-manager/blob/main/installation/ersm_setup.exe)

> [!WARN]
> If you're (rightfully) concerned about downloading an executable from GitHub, it should be relatively straightforward to make your own installer with [InstallForge](https://installforge.net/).

Just build the executable with
```console
go build -ldflags -H=windowsgui src/main.go src/save-management.go src/config.go
```
and follow the basic steps to create the installer with your executable.

If you want your build to have icons, make sure to use [icon64.png](icon64.png) in the assets folder in InstallForge.

My InstallForge save file is in the installation folder if you want to use it as a reference.


## Setup

1. Ensure that the save file loaded in Elden Ring is what you want to be your root save file.
Your root save file should be the one with the most weapons, items, or larval tears.
2. Create a folder to store your save files.
3. Start Elden Ring Save Manager. Once you do this you will be prompted to give a path to your save file folder. Paste in the path and click Save.

### Setup for users that already manage their save files

If you already have a folder with your save files setup, rename your root save file to `ROOT`
and make sure that each save has the following files:
- **`ER0000.sl2`**
- **`ER0000.sl2.bak`**
- **`steam_autocloud.vdf`**

Ensure that the save file loaded in Elden Ring is what you want to be your root save file.

Then, when you start the program for the first time, paste the path to the folder with your save files and click Save
