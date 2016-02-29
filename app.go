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
	http.HandleFunc("/init", initHandler)
	http.HandleFunc("/generate", generateHandler)
}

func initHandler(w http.ResponseWriter, r *http.Request) {
	json, err := command.InitResumeJson()
	if err != nil {
		errorHandler(w, r, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\"resume.json\"")
	w.Write([]byte(json))
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Update JSON-Resume data file
	const MAX_UPLOAD_BYTES = 100000
	file, _, err := r.FormFile("file")
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("An error occurred uploading the file: %s\n", err))
		return
	}
	ctx.Infof("file == %s\n", file)
	uploadBuffer := make([]byte, MAX_UPLOAD_BYTES)
	bytesRead, _ := file.Read(uploadBuffer)
	if bytesRead >= MAX_UPLOAD_BYTES {
		errorHandler(w, r, fmt.Sprintf("JSON Resume upload exceeds the %d byte cap\n", MAX_UPLOAD_BYTES))
		return
	}
	ctx.Infof("Uploaded %d bytes of JSON-Resume data\n", bytesRead)
	contents := string(uploadBuffer[:bytesRead])

	// Parse the (hopefully) JSON
	resumeData, err := data.FromJsonString(contents)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("An error occurred parsing the uploaded file: %s\n", err))
		return
	}

	// Load the selected template
	templateParam := r.Form.Get("template")
	if templateParam != "standard" && templateParam != "iconic" && templateParam != "refined" {
		errorHandler(w, r, fmt.Sprintf("An unrecognized template was selected: %s\n", templateParam))
		return
	}
	ctx.Infof("Exporting a resume with template: %s\n", templateParam)
	templateBytes, err := ioutil.ReadFile(filepath.Join("templates", templateParam+".xml"))
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("An error occurred loading the template file: %s\n", err))
		return
	}
	templateString := string(templateBytes)

	// Generate the resume
	exportBuffer, err := command.ExportResume(resumeData, templateString)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("An error occurred exporting the resume: %s\n", err))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\"resume.doc\"")
	w.Write(exportBuffer.Bytes())
}

func errorHandler(w http.ResponseWriter, r *http.Request, message string) {
	ctx := appengine.NewContext(r)
	ctx.Errorf("Handling error: %s\n", message)

	w.Header().Set("Content-Type", "text/html")
	t, err := template.ParseFiles(path.Join("static", "error.html"))
	if err != nil {
		ctx.Errorf("Error loading 'error.html': %s\n", err)
		w.Write([]byte("Unable to render the error page"))
		return
	}
	data := make(map[string]string)
	data["message"] = message
	t.Execute(w, data)
}

