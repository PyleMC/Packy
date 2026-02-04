package app

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window
}

func New() *App {
	a := fyneApp.NewWithID("com.pylemc.packy")
	w := a.NewWindow("Packy")
	w.SetMaster()

	content := container.NewStack()
	zipView := buildZipView()
	encryptView := buildEncryptView()
	decryptView := buildDecryptView()
	content.Objects = []fyne.CanvasObject{zipView}

	title := canvas.NewText("Packy", a.Settings().Theme().Color(theme.ColorNameForeground, a.Settings().ThemeVariant()))
	title.TextSize = 24

	nav := buildNav(func(name string) {
		switch name {
		case "Zip Pack":
			content.Objects = []fyne.CanvasObject{zipView}
		case "Encrypt Pack":
			content.Objects = []fyne.CanvasObject{encryptView}
		case "Decrypt Pack":
			content.Objects = []fyne.CanvasObject{decryptView}
		default:
			content.Objects = []fyne.CanvasObject{zipView}
		}
		content.Refresh()
	})

	sidebar := container.NewBorder(title, nil, nil, nil, nav)
	sidebarWrap := container.New(newFixedSizeLayoutExpand(fyne.NewSize(220, 0)), sidebar)
	split := container.NewBorder(nil, nil, sidebarWrap, nil, content)
	w.SetContent(split)
	w.Resize(fyne.NewSize(780, 520))

	return &App{
		fyneApp: a,
		window:  w,
	}
}

func (a *App) Run() {
	a.window.ShowAndRun()
}

func buildNav(onSelect func(name string)) fyne.CanvasObject {
	items := []string{"Zip Pack", "Encrypt Pack", "Decrypt Pack"}
	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			if uid == "" {
				return items
			}
			return []string{}
		},
		IsBranch: func(uid string) bool {
			return uid == ""
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(uid)
		},
		OnSelected: func(uid string) {
			if uid != "" {
				onSelect(uid)
			}
		},
	}
	tree.Select("Zip Pack")
	return tree
}

func buildZipView() fyne.CanvasObject {
	status := widget.NewLabel("No actions yet.")
	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("Path to resource pack folder")
	outputEntry := widget.NewEntry()
	outputEntry.SetPlaceHolder("Output zip path (optional)")

	var zipBtn *widget.Button
	zipBtn = widget.NewButton("Zip resource pack", func() {
		runAction(status, zipBtn, func() (string, error) {
			folderPath := strings.TrimSpace(folderEntry.Text)
			outputPath := strings.TrimSpace(outputEntry.Text)
			if folderPath == "" {
				return "", fmt.Errorf("folder path is required")
			}
			if err := validateResourcePackFolder(folderPath); err != nil {
				return "", err
			}
			zipPath, err := zipPackFromFolder(folderPath, outputPath)
			if err != nil {
				return "", err
			}
			return "Saved zip: " + zipPath, nil
		})
	})

	note := widget.NewLabel("Zips the contents of the resourcepack folder into a .zip.")
	form := widget.NewForm(
		widget.NewFormItem("Folder Path", folderEntry),
		widget.NewFormItem("Output Zip", outputEntry),
	)
	return container.NewBorder(nil, status, nil, nil, container.NewVBox(note, form, zipBtn))
}

func buildEncryptView() fyne.CanvasObject {
	status := widget.NewLabel("No actions yet.")
	zipEntry := widget.NewEntry()
	zipEntry.SetPlaceHolder("Path to zip pack")
	keyEntry := widget.NewPasswordEntry()
	keyEntry.SetPlaceHolder("Encryption key/password")
	outputEntry := widget.NewEntry()
	outputEntry.SetPlaceHolder("Output encrypted path (optional)")

	var encryptBtn *widget.Button
	encryptBtn = widget.NewButton("Encrypt zip pack", func() {
		runAction(status, encryptBtn, func() (string, error) {
			return "Encrypt action not implemented.", nil
		})
	})

	form := widget.NewForm(
		widget.NewFormItem("Zip Path", zipEntry),
		widget.NewFormItem("Key", keyEntry),
		widget.NewFormItem("Output", outputEntry),
	)
	return container.NewBorder(nil, status, nil, nil, container.NewVBox(form, encryptBtn))
}

func buildDecryptView() fyne.CanvasObject {
	status := widget.NewLabel("No actions yet.")
	encryptedEntry := widget.NewEntry()
	encryptedEntry.SetPlaceHolder("Path to encrypted pack")
	keyEntry := widget.NewPasswordEntry()
	keyEntry.SetPlaceHolder("Decryption key/password")
	outputEntry := widget.NewEntry()
	outputEntry.SetPlaceHolder("Output zip path (optional)")

	var decryptBtn *widget.Button
	decryptBtn = widget.NewButton("Decrypt pack", func() {
		runAction(status, decryptBtn, func() (string, error) {
			return "Decrypt action not implemented.", nil
		})
	})

	form := widget.NewForm(
		widget.NewFormItem("Encrypted Path", encryptedEntry),
		widget.NewFormItem("Key", keyEntry),
		widget.NewFormItem("Output", outputEntry),
	)
	return container.NewBorder(nil, status, nil, nil, container.NewVBox(form, decryptBtn))
}

type fixedSizeLayoutExpand struct {
	size fyne.Size
}

func newFixedSizeLayoutExpand(size fyne.Size) fyne.Layout {
	return &fixedSizeLayoutExpand{size: size}
}

func (l *fixedSizeLayoutExpand) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	width := l.size.Width
	height := l.size.Height
	if width <= 0 {
		width = size.Width
	}
	if height <= 0 {
		height = size.Height
	}
	for _, obj := range objects {
		obj.Move(fyne.NewPos(0, 0))
		obj.Resize(fyne.NewSize(width, height))
	}
}

func (l *fixedSizeLayoutExpand) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return l.size
}

func runAction(status *widget.Label, button *widget.Button, action func() (string, error)) {
	button.Disable()
	status.SetText("Working...")
	go func() {
		message, err := action()
		if err != nil {
			status.SetText(err.Error())
		} else if message != "" {
			status.SetText(message)
		} else {
			status.SetText("Done.")
		}
		button.Enable()
	}()
}
