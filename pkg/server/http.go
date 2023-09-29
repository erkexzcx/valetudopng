package server

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/erkexzcx/valetudopng"
)

func runWebServer(bind string) {
	http.HandleFunc("/api/map/image", requestHandlerImage)
	http.HandleFunc("/api/map/image/debug", requestHandlerDebug)
	http.HandleFunc("/api/map/image/debug/static/", requestHandlerDebugStatic)
	panic(http.ListenAndServe(bind, nil))
}

func isResultNotReady() bool {
	renderedPNGMux.RLock()
	defer renderedPNGMux.RUnlock()
	return result == nil
}

func requestHandlerImage(w http.ResponseWriter, r *http.Request) {
	if isResultNotReady() {
		http.Error(w, "image not yet loaded", http.StatusAccepted)
		return
	}

	renderedPNGMux.RLock()
	imageCopy := make([]byte, len(renderedPNG))
	copy(imageCopy, renderedPNG)
	renderedPNGMux.RUnlock()

	w.Header().Set("Content-Length", strconv.Itoa(len(imageCopy)))
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(200)
	w.Write(imageCopy)
}

type TemplateData struct {
	RobotMinX    int
	RobotMinY    int
	RobotMaxX    int
	RobotMaxY    int
	RotatedTimes int
	Scale        int
	PixelSize    int
}

func requestHandlerDebug(w http.ResponseWriter, r *http.Request) {
	if isResultNotReady() {
		http.Error(w, "image not yet loaded", http.StatusAccepted)
		return
	}

	// Parse the template file
	tmpl, err := template.ParseFS(valetudopng.WebFS, "web/templates/index.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Create a data structure to hold the template values
	renderedPNGMux.RLock()
	data := TemplateData{
		RobotMinX:    result.RobotCoords.MinX,
		RobotMinY:    result.RobotCoords.MinY,
		RobotMaxX:    result.RobotCoords.MaxX,
		RobotMaxY:    result.RobotCoords.MaxY,
		RotatedTimes: result.Settings.RotationTimes,
		Scale:        int(result.Settings.Scale),
		PixelSize:    result.PixelSize,
	}
	renderedPNGMux.RUnlock()

	// Render the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func requestHandlerDebugStatic(w http.ResponseWriter, r *http.Request) {
	staticPath := "web/static/" + strings.TrimPrefix(r.URL.Path, "/api/map/image/debug/static/")
	file, err := valetudopng.WebFS.Open(staticPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Read the entire file into memory
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create an io.ReadSeeker from the byte slice
	reader := bytes.NewReader(data)

	http.ServeContent(w, r, info.Name(), info.ModTime(), reader)
}
