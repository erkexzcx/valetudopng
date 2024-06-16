package server

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/erkexzcx/valetudopng"
)

func runWebServer(ctx context.Context, wg *sync.WaitGroup, panic chan bool, bind string) {

	http.HandleFunc("/api/map/image", requestHandlerImage)
	http.HandleFunc("/api/map/image/debug", requestHandlerDebug)
	http.HandleFunc("/api/map/image/debug/lovelace/", requestHandlerDebugConfig)
	http.HandleFunc("/api/map/image/debug/static/", requestHandlerDebugStatic)
	server := http.Server{
		Addr:    bind,
		Handler: http.DefaultServeMux,
	}

	go func() {
	DONE:
		for {
			select {
			case <-ctx.Done():
				break DONE
			case <-panic:
				break DONE
			}
		}
		server.Shutdown(context.Background())
		wg.Done()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", slog.String("error", err.Error()))
		panic <- true
	}
	slog.Info("HTTP server shut down")
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

func requestHandlerDebugConfig(w http.ResponseWriter, _ *http.Request) {
	if isResultNotReady() {
		http.Error(w, "image not yet loaded", http.StatusAccepted)
		return
	}

	// TODO: add lock
	w.Header().Set("Content-Length", strconv.Itoa(len(renderedCfg)))
	//	w.Header().Set("Content-Type", "application/x-yaml") // is this preferred? annoying when the browser downloads instead of shows
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(200)
	renderedPNGMux.RLock()
	defer renderedPNGMux.RUnlock()
	w.Write(renderedCfg)
}
