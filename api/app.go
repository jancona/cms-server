package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jancona/ourroots/model"
)

const contentType = "application/json"

// App is the container for the application
type App struct {
	categoryPersister   model.CategoryPersister
	collectionPersister model.CollectionPersister
	baseURL             url.URL
	validate            *validator.Validate
}

// NewApp builds an App
func NewApp() *App {
	validate := validator.New()
	// Return JSON tag name as Field() in errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &App{
		baseURL:  url.URL{},
		validate: validate,
	}
}

// BaseURL sets the base URL for the app
func (app *App) BaseURL(url url.URL) *App {
	app.baseURL = url
	return app
}

// Validate sets the validate object for the app
func (app *App) Validate(validate *validator.Validate) *App {
	app.validate = validate
	return app
}

// CategoryPersister sets the CategoryPersister for the app
func (app *App) CategoryPersister(cp model.CategoryPersister) *App {
	app.categoryPersister = cp
	return app
}

// CollectionPersister sets the CollectionPersister for the app
func (app *App) CollectionPersister(cp model.CollectionPersister) *App {
	app.collectionPersister = cp
	return app
}

// GetIndex redirects to the Swagger documentation
func (app App) GetIndex(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, app.baseURL.Path+"/swagger/", http.StatusTemporaryRedirect)
}

// NewRouter builds a router for handling requests
func (app App) NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc(app.baseURL.Path+"/", app.GetIndex).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/index.html", app.GetIndex).Methods("GET")

	r.HandleFunc(app.baseURL.Path+"/categories", app.GetAllCategories).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/categories", app.PostCategory).Methods("POST")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.GetCategory).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.PatchCategory).Methods("PATCH")
	r.HandleFunc(app.baseURL.Path+"/categories/{id}", app.DeleteCategory).Methods("DELETE")

	r.HandleFunc(app.baseURL.Path+"/collections", app.GetAllCollections).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/collections", app.PostCollection).Methods("POST")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.GetCollection).Methods("GET")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.PatchCollection).Methods("PATCH")
	r.HandleFunc(app.baseURL.Path+"/collections/{id}", app.DeleteCollection).Methods("DELETE")
	return r
}

// NotFound returns an http.StatusNotFound response
func NotFound(w http.ResponseWriter, req *http.Request) {
	m := fmt.Sprintf("Path '%s' not found", req.URL.RequestURI())
	log.Print("[ERROR] " + m)
	OtherErrorResponse(w, http.StatusNotFound, m)
}

func serverError(w http.ResponseWriter, err error) {
	log.Print("[ERROR] Server error: " + err.Error())
	// debug.PrintStack()
	OtherErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err.Error()))
}

// OtherErrorResponse returns an error response
func OtherErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	e := model.NewError(model.ErrOther, message)
	err := enc.Encode(model.Errors{e})
	if err != nil {
		log.Printf("[ERROR] Failure encoding error response: '%v'", err)
	}
}

// ValidationErrorResponse returns a validation error response
func ValidationErrorResponse(w http.ResponseWriter, code int, er error) {
	// func ValidationErrorResponse(w http.ResponseWriter, code int, errorCode model.ErrorCode, params ...string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	errors := make(model.Errors, 0)
	if ves, ok := er.(validator.ValidationErrors); ok {
		for _, fe := range ves {
			if fe.Tag() == "required" {
				name := strings.SplitN(fe.Namespace(), ".", 2)
				// log.Printf("name: %v", name)
				errors = append(errors, model.NewError(model.ErrRequired, name[1]))
			} else {
				errors = append(errors, model.NewError(model.ErrOther, fmt.Sprintf("Key: '%s' Error:Field validation for '%s' failed on the '%s' tag", fe.Namespace(), fe.Field(), fe.Tag())))
			}
		}
	} else {
		errors = append(errors, model.NewError(model.ErrOther, er.Error()))
	}
	log.Printf("[DEBUG] errBody: %#v", errors)
	err := enc.Encode(errors)
	if err != nil {
		log.Printf("[ERROR] Failure encoding error response: '%v'", err)
	}
}
