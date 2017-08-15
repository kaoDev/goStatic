// This small program is just a small web server created in static mode
// in order to provide the smallest docker image possible

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	// Def of flags
	portPtr    = flag.Int("p", 8043, "The listening port")
	root       = flag.String("static", "./srv/http", "The path for the static files")
	indexPath  = flag.String("staticIndex", "./srv/http/index.html", "The path for the static index file")
	headerFlag = flag.String("appendHeader", "", "HTTP response header, specified as `HeaderName:Value` that should be added to all responses.")
)

func parseHeaderFlag(headerFlag string) (string, string) {
	if len(headerFlag) == 0 {
		return "", ""
	}
	pieces := strings.SplitN(headerFlag, ":", 2)
	if len(pieces) == 1 {
		return pieces[0], ""
	}
	return pieces[0], pieces[1]
}

type customFileServer struct {
	root            http.Dir
	NotFoundHandler func(http.ResponseWriter, *http.Request)
}

func CustomFileServer(root http.Dir, NotFoundHandler http.HandlerFunc) http.Handler {
	return &customFileServer{root: root, NotFoundHandler: NotFoundHandler}
}

func isSlashRune(r rune) bool {
	return r == '/' || r == '\\'
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func (fs *customFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//if empty, set current directory
	dir := string(fs.root)
	if dir == "" {
		dir = "."
	}

	//add prefix and clean
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	upath = path.Clean(upath)

	//path to file
	name := path.Join(dir, filepath.FromSlash(upath))

	//check if file exists
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			fs.NotFoundHandler(w, r)
			return
		}
	}
	defer f.Close()

	http.ServeFile(w, r, name)
}

func main() {

	flag.Parse()

	port := ":" + strconv.FormatInt(int64(*portPtr), 10)

	serveIndexHtml := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, *indexPath)
	}

	mux := http.NewServeMux()
	mux.Handle("/", CustomFileServer(http.Dir(*root), serveIndexHtml))

	server := &http.Server{
		Addr:           port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening at 0.0.0.0%v...", port)
	log.Fatalln(server.ListenAndServe())
}
