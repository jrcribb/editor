package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

type Desktop struct {
	ctx        context.Context
	allowClose atomic.Bool
	mu         sync.Mutex
	arguments  []string
	paths      map[int]string
}

type FileHandle struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

type FileInfo struct {
	Size         int64 `json:"size"`
	LastModified int64 `json:"lastModified"`
}

type OpenedFiles struct {
	Files  []FileHandle `json:"files"`
	Errors []string     `json:"errors"`
}

var textFilters = []runtime.FileFilter{
	{DisplayName: "Markdown y texto", Pattern: "*.md;*.markdown;*.mdx;*.txt;*.text"},
	{DisplayName: "Todos los archivos", Pattern: "*.*"},
}

func (d *Desktop) StartupFiles() OpenedFiles {
	d.mu.Lock()
	args := d.arguments
	d.arguments = nil
	d.mu.Unlock()

	var paths, problems []string
	for _, arg := range args {
		if !strings.ContainsAny(filepath.Base(arg), "*?") {
			paths = append(paths, arg)
			continue
		}
		matches, err := matchWindowsFiles(arg)
		if err != nil {
			problems = append(problems, arg+": "+err.Error())
		} else if len(matches) == 0 {
			problems = append(problems, arg+": no se encontraron archivos")
		}
		paths = append(paths, matches...)
	}
	result := d.openPaths(paths)
	result.Errors = append(problems, result.Errors...)
	seen := make(map[int]bool)
	for _, file := range result.Files {
		seen[file.ID] = true
	}
	for _, path := range loadSessionPaths() {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		file, err := d.addPath(path)
		if err == nil && !seen[file.ID] {
			result.Files = append(result.Files, file)
			seen[file.ID] = true
		}
	}
	return result
}

func sessionPath() (string, error) {
	dir, err := os.UserConfigDir()
	return filepath.Join(dir, "MarkdownEditor", "session.json"), err
}

func loadSessionPaths() []string {
	path, err := sessionPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var paths []string
	if json.Unmarshal(data, &paths) != nil {
		return nil
	}
	return paths
}

func (d *Desktop) SaveSession(ids []int) error {
	paths := make([]string, 0, len(ids))
	for _, id := range ids {
		if path, err := d.path(id); err == nil {
			paths = append(paths, path)
		}
	}
	path, err := sessionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), "session-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(paths)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}

func matchWindowsFiles(pattern string) ([]string, error) {
	pattern, err := filepath.Abs(pattern)
	if err != nil {
		return nil, err
	}
	search, err := windows.UTF16PtrFromString(pattern)
	if err != nil {
		return nil, err
	}
	var data windows.Win32finddata
	handle, err := windows.FindFirstFile(search, &data)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer windows.FindClose(handle)

	dir := filepath.Dir(pattern)
	var matches []string
	for err == nil {
		if data.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
			matches = append(matches, filepath.Join(dir, windows.UTF16ToString(data.FileName[:])))
		}
		err = windows.FindNextFile(handle, &data)
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

func (d *Desktop) OpenFiles() (OpenedFiles, error) {
	paths, err := runtime.OpenMultipleFilesDialog(d.ctx, runtime.OpenDialogOptions{
		Title: "Abrir archivos", Filters: textFilters,
	})
	return d.openPaths(paths), err
}

func (d *Desktop) openPaths(paths []string) OpenedFiles {
	var result OpenedFiles
	for _, path := range paths {
		handle, err := d.addPath(path)
		if err != nil {
			result.Errors = append(result.Errors, path+": "+err.Error())
			continue
		}
		result.Files = append(result.Files, handle)
	}
	return result
}

func (d *Desktop) SaveTarget(name string) (*FileHandle, error) {
	path, err := runtime.SaveFileDialog(d.ctx, runtime.SaveDialogOptions{
		Title: "Guardar como", DefaultFilename: filepath.Base(name),
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown (*.md)", Pattern: "*.md"},
			{DisplayName: "Todos los archivos", Pattern: "*.*"},
		},
	})
	if err != nil || path == "" {
		return nil, err
	}
	handle, err := d.addPath(path)
	return &handle, err
}

func (d *Desktop) addPath(path string) (FileHandle, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return FileHandle{}, err
	}
	name := filepath.Base(path)
	folder := filepath.Dir(path)
	// Resolve aliases for existing files so replacing a file keeps a symlink intact.
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for id, existing := range d.paths {
		if existing == path {
			return FileHandle{ID: id, Name: name, Folder: folder}, nil
		}
	}
	id := len(d.paths) + 1
	d.paths[id] = path
	return FileHandle{ID: id, Name: name, Folder: folder}, nil
}

func (d *Desktop) path(id int) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	path, ok := d.paths[id]
	if !ok {
		return "", errors.New("archivo desconocido")
	}
	return path, nil
}

func (d *Desktop) Stat(id int) (FileInfo, error) {
	path, err := d.path(id)
	if err != nil {
		return FileInfo{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	if !info.Mode().IsRegular() {
		return FileInfo{}, errors.New("la ruta no corresponde a un archivo")
	}
	return FileInfo{info.Size(), info.ModTime().UnixMilli()}, nil
}

func (d *Desktop) Read(id int) (string, error) {
	path, err := d.path(id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return "", errors.New("el archivo no es texto UTF-8")
	}
	return string(data), nil
}

func (d *Desktop) Write(id int, text string) error {
	path, err := d.path(id)
	if err != nil {
		return err
	}
	return writeEditorFile(path, []byte(text))
}

func (d *Desktop) SavePDF(name, encoded string) error {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || !bytes.HasPrefix(data, []byte("%PDF-")) {
		return errors.New("el PDF generado no es válido")
	}
	path, err := runtime.SaveFileDialog(d.ctx, runtime.SaveDialogOptions{
		Title: "Guardar PDF", DefaultFilename: filepath.Base(name),
		Filters: []runtime.FileFilter{{DisplayName: "Documento PDF (*.pdf)", Pattern: "*.pdf"}},
	})
	if err != nil || path == "" {
		return err
	}
	if filepath.Ext(path) == "" {
		path += ".pdf"
	}
	return writeEditorFile(path, data)
}

func writeEditorFile(path string, data []byte) error {
	// Stage the complete edit alongside the destination before replacing it.
	file, err := os.CreateTemp(filepath.Dir(path), ".markdown-editor-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}
