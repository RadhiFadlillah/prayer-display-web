package backend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

var developmentMode = false

// ServeApp serves web app in specified port
func ServeApp(port int) error {
	// Create router
	router := httprouter.New()
	router.GET("/", serveIndex)
	router.GET("/image/:name", serveImage)
	router.GET("/js/*filepath", serveJsFile)
	router.GET("/res/*filepath", serveFile)
	router.GET("/css/*filepath", serveFile)
	router.GET("/api/data", loadData)

	// Route for panic
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
		http.Error(w, fmt.Sprint(arg), 500)
	}

	// Create server
	url := fmt.Sprintf(":%d", port)
	svr := &http.Server{
		Addr:         url,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: time.Minute,
	}

	// Serve app
	logrus.Infoln("Serve app in", url)
	return svr.ListenAndServe()
}
