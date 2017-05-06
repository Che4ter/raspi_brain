package server

//https://github.com/jmcfarlane/golang-templates-example/blob/master/main.go
//go-bindata-assetfs ../static/... ../templates/...

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/configuration"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Model of stuff to render a page
type MainSiteModel struct {
	Title       string
	MainHeading string
	Body        string
}

// Templates with functions available to them
var templates = template.New("").Funcs(templateMap)

// Parse all of the bindata templates
func init() {
	for _, path := range AssetNames() {
		bytes, err := Asset(path)
		if err != nil {
			log.Panicf("Unable to parse: path=%s, err=%s", path, err)
		}
		templates.New(path).Parse(string(bytes))
	}
}

var (
	templateMap = template.FuncMap{
		"Upper": func(s string) string {
			return strings.ToUpper(s)
		},
	}
)

// Render a template given a model
func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	model := MainSiteModel{Title: "Raspberry PI Configuration UI", MainHeading: "Pren Team 9 - Caterpillar", Body: "Hello World"}
	renderTemplate(w, "templates/index.html", &model)
}

func saveHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	/*title := r.URL.Path[len("/save/"):]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	p.save()*/
	http.Redirect(w, r, "/", http.StatusFound)
}

func RunWebServer(chDefault chan float32, config configuration.Configuration) {
	// mux handler
	router := httprouter.New()

	// Example route that takes one rest style option
	router.GET("/", index)

	router.POST("/save", saveHandler)

	// Serve static assets via the "static" directory
	router.ServeFiles("/static/*filepath", assetFS())

	// Serve this program forever
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.WebPort), router))
	fmt.Println("running web iterface on port:", config.WebPort)

	//http.HandleFunc("/save/", saveHandler)
}
