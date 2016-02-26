package ResumeFodder

import (
	"appengine"
	"gitlab.com/steve-perkins/ResumeFodder/data"
	"html/template"
	"net/http"
	"path"
)

func init() {
	http.HandleFunc("/test", testHandler)
	http.HandleFunc("/init", initHandler)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(path.Join("html", "test.html"))
	if err != nil {
		ctx := appengine.NewContext(r)
		ctx.Errorf("Error loading 'test.html': %s\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Unable to serve 'test.html'"))
	}
	t.Execute(w, nil)
}

func initHandler(w http.ResponseWriter, r *http.Request) {
	json, err := data.ToJsonString(data.NewResumeData())
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(json))
}
