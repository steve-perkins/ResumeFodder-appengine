package ResumeFodder

import (
	"appengine"
	"net/http"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/foo", defaultHandler)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx.Infof("Hello world\n")
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("Hello world"))
}
