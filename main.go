package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/oxplot/pdfrankestein/session"
)

const (
	progName = "PDFrankestein"
)

func run() error {

	ap := app.New()
	win := ap.NewWindow(progName)

	var sess *session.Session
	sig := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		select {
		case <-sig:
			break
		case <-done:
			break
		}
		if sess != nil {
			sess.Close()
		}
		win.Close()
	}()

	fileNameLabel := widget.NewLabel("abc.pdf")
	filePathLabel := widget.NewLabel("/home/...")

	var openedContent *fyne.Container

	var pages []*widget.Button
	pageGrid := container.NewGridWrap(fyne.NewSize(100, 100))

	editingMsg := container.NewCenter(widget.NewLabel("Annotating in Inkscape"))

	startContent := container.NewCenter(widget.NewButton("Open PDF File", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				dialog.ShowError(err, win)
				return
			}
			r.Close()
			path := r.URI().String()
			if !strings.HasPrefix(path, "file://") {
				dialog.ShowError(fmt.Errorf("invalid file selected"), win)
				return
			}
			path = strings.TrimPrefix(path, "file://")

			sess, err = session.New(path)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			pages = make([]*widget.Button, sess.PageCount())
			for i := range pages {
				p := func(page int) *widget.Button {
					return widget.NewButton("p."+strconv.Itoa(page+1), func() {
						win.SetContent(editingMsg)
						_, err := sess.Annotate(page)
						if err != nil {
							dialog.ShowError(err, win)
						}
						win.SetContent(openedContent)
					})
				}(i)
				pages[i] = p
				pageGrid.Add(p)
			}
			fileNameLabel.SetText(r.URI().Name())
			filePathLabel.SetText(path)
			win.SetContent(openedContent)

		}, win)
	}))

	openedContent = container.NewBorder(
		container.NewBorder(
			nil, nil, nil,
			container.NewHBox(
				widget.NewButton("Save", func() {
				}),
				widget.NewButton("Close", func() {
					pageGrid.Objects = nil
					sess.Close()
					sess = nil
					win.SetContent(startContent)
				}),
			),
			container.NewVBox(
				fileNameLabel,
				filePathLabel,
			),
		),
		nil, nil, nil,
		container.NewVScroll(pageGrid),
	)

	win.Resize(fyne.NewSize(600, 500))
	win.SetContent(startContent)
	win.ShowAndRun()

	return nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("pdfrankestein: ")
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
