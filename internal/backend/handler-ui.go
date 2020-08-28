package backend

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

var presetMimeTypes = map[string]string{
	".css":  "text/css; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".js":   "application/javascript",
	".png":  "image/png",
}

// serveFile serves general UI file
func serveFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := serveAssets(w, r, r.URL.Path)
	checkError(err)
}

// serveJsFile serves all JS file
func serveJsFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsFilePath := ps.ByName("filepath")
	jsFilePath = path.Join("js", jsFilePath)
	jsDir, jsName := path.Split(jsFilePath)

	if developmentMode && fp.Ext(jsName) == ".js" && strings.HasSuffix(jsName, ".min.js") {
		jsName = strings.TrimSuffix(jsName, ".min.js") + ".js"
		tmpPath := path.Join(jsDir, jsName)
		if assetExists(tmpPath) {
			jsFilePath = tmpPath
		}
	}

	err := serveAssets(w, r, jsFilePath)
	checkError(err)
}

// serveIndex serves the index page
func serveIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := serveAssets(w, r, "index.html")
	checkError(err)
}

func serveAssets(w http.ResponseWriter, r *http.Request, filePath string) error {
	// Open file
	src, err := assets.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Get file statistic
	info, err := src.Stat()
	if err != nil {
		return err
	}

	// Get content type
	ext := fp.Ext(filePath)
	mimeType := guessTypeByExtension(ext)

	// Write response header
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}

	// Serve file
	_, err = io.Copy(w, src)
	return err
}

func guessTypeByExtension(ext string) string {
	ext = strings.ToLower(ext)

	if v, ok := presetMimeTypes[ext]; ok {
		return v
	}

	return mime.TypeByExtension(ext)
}

func assetExists(filePath string) bool {
	f, err := assets.Open(filePath)
	if f != nil {
		f.Close()
	}
	return err == nil || !os.IsNotExist(err)
}
