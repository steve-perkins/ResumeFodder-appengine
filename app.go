package ResumeFodder

import (
	"appengine"
	"fmt"
	"gitlab.com/steve-perkins/ResumeFodder/data"
	"html/template"
	"net/http"
	"path"
)

func init() {
	http.HandleFunc("/test", testHandler)
	http.HandleFunc("/init", initHandler)
	http.HandleFunc("/generate", generateHandler)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(path.Join("html", "test.html"))
	if err != nil {
		ctx := appengine.NewContext(r)
		ctx.Errorf("Error loading 'test.html': %s\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Unable to serve 'test.html'"))
		return
	}
	t.Execute(w, nil)
}

func initHandler(w http.ResponseWriter, r *http.Request) {
	json, err := data.ToJsonString(data.NewResumeData())
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(json))
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	// Update JSON-Resume data file
	const MAX_UPLOAD_BYTES = 100000
	file, _, _ := r.FormFile("file")
	buffer := make([]byte, MAX_UPLOAD_BYTES)
	bytesRead, _ := file.Read(buffer)
	ctx := appengine.NewContext(r)
	if bytesRead >= MAX_UPLOAD_BYTES {
		ctx.Errorf("JSON-Resume upload exceeds the %d byte cap\n", MAX_UPLOAD_BYTES)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf("Error: File uploads cannot exceed %d bytes\n", MAX_UPLOAD_BYTES)))
		return
	}
	ctx.Infof("Uploaded %d bytes of JSON-Resume data\n", bytesRead)
	contents := string(buffer[:bytesRead])

	// Parse the (hopefully) JSON
	_, err := data.FromJsonString(contents)
	if err != nil {
		ctx.Errorf("An error occurred parsing the uploaded file: %s\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Cannot parse the uploaded file as JSON-Resume data\n"))
		return
	}

	// Generate the resume
	// TODO: Refactor the base ResumeFodder project, to provide an export method that works with in-memory strings rather than the filesystem

	// TODO: Remove this filler
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(contents))
}
