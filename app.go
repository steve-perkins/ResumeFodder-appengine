package ResumeFodder

import (
	"appengine"
	"fmt"
	"gitlab.com/steve-perkins/ResumeFodder/command"
	"gitlab.com/steve-perkins/ResumeFodder/data"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir("static")))
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
	json, err := command.InitResumeJson()
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\"resume.json\"")
	w.Write([]byte(json))
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	// Update JSON-Resume data file
	const MAX_UPLOAD_BYTES = 100000
	file, _, _ := r.FormFile("file")
	uploadBuffer := make([]byte, MAX_UPLOAD_BYTES)
	bytesRead, _ := file.Read(uploadBuffer)
	ctx := appengine.NewContext(r)
	if bytesRead >= MAX_UPLOAD_BYTES {
		ctx.Errorf("JSON-Resume upload exceeds the %d byte cap\n", MAX_UPLOAD_BYTES)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf("Error: File uploads cannot exceed %d bytes\n", MAX_UPLOAD_BYTES)))
		return
	}
	ctx.Infof("Uploaded %d bytes of JSON-Resume data\n", bytesRead)
	contents := string(uploadBuffer[:bytesRead])

	// Parse the (hopefully) JSON
	resumeData, err := data.FromJsonString(contents)
	if err != nil {
		ctx.Errorf("An error occurred parsing the uploaded file: %s\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Cannot parse the uploaded file as JSON-Resume data\n"))
		return
	}

	// Load the selected template
	templateParam := r.Form.Get("template")
	if templateParam != "plain" && templateParam != "iconic" && templateParam != "refined" {
		ctx.Errorf("An unrecognized template was selected: %s\n", templateParam)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Cannot load the template\n"))
		return
	}
	ctx.Infof("Exporting a resume with template: %s\n", templateParam)
	templateBytes, err := ioutil.ReadFile(filepath.Join("templates", templateParam+".xml"))
	if err != nil {
		ctx.Errorf("An error occurred loading the template file: %s\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Cannot load the template\n"))
		return
	}
	templateString := string(templateBytes)

	// Generate the resume
	exportBuffer, err := command.ExportResume(resumeData, templateString)
	if err != nil {
		ctx.Errorf("An error occurred exporting the resume: %s\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("An error occurred exporting the resume\n"))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\"resume.doc\"")
	w.Write(exportBuffer.Bytes())
}
