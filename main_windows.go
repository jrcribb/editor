package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

//go:embed index.html
var assets embed.FS

const editorTitle = "Editor"

func main() {
	editor := &Desktop{paths: make(map[int]string), arguments: os.Args[1:]}
	err := wails.Run(&options.App{
		Title:     editorTitle,
		Width:     1160,
		Height:    780,
		MinWidth:  360,
		MinHeight: 320,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// The document renders untrusted Markdown, while the app itself uses inline scripts.
					w.Header().Set("Content-Security-Policy", "script-src 'self' 'unsafe-inline'; frame-src 'none'; object-src 'none'; base-uri 'none'; form-action 'none'")
					next.ServeHTTP(w, r)
				})
			},
		},
		DragAndDrop: &options.DragAndDrop{EnableFileDrop: true},
		OnStartup: func(ctx context.Context) {
			editor.ctx = ctx
			runtime.OnFileDrop(ctx, func(_, _ int, paths []string) {
				runtime.EventsEmit(ctx, "editor:drop", editor.openPaths(paths))
			})
		},
		OnBeforeClose: func(ctx context.Context) bool {
			if editor.allowClose.Load() {
				return false
			}
			runtime.EventsEmit(ctx, "editor:close")
			return true
		},
		Bind: []interface{}{editor},
	})
	if err != nil {
		title, _ := windows.UTF16PtrFromString(editorTitle)
		message, _ := windows.UTF16PtrFromString(fmt.Sprintf("No se pudo iniciar el editor. Comprueba que Microsoft Edge WebView2 Runtime esté instalado.\n\n%v", err))
		windows.MessageBox(0, message, title, windows.MB_OK|windows.MB_ICONERROR)
		os.Exit(1)
	}
}

func (d *Desktop) ShowMessage(message string, question bool) (bool, error) {
	dialog := runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   editorTitle,
		Message: message,
	}
	if question {
		dialog.Type = runtime.QuestionDialog
		dialog.DefaultButton = "No"
	}
	response, err := runtime.MessageDialog(d.ctx, dialog)
	return response == "Yes", err
}

// CloseWindow completes the close request after the shared UI checks for unsaved edits.
func (d *Desktop) CloseWindow() {
	d.allowClose.Store(true)
	runtime.Quit(d.ctx)
}
