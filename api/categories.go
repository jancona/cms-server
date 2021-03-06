package api

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/jancona/ourroots/model"
	"github.com/jancona/ourroots/persist"
)

// GetAllCategories returns all categories in the database
// @summary returns all categories
// @router /categories [get]
// @tags categories
// @id getCategories
// @produce application/json
// @success 200 {array} model.Category "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetAllCategories(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cats, err := app.categoryPersister.SelectCategories()
	if err != nil {
		serverError(w, err)
		return
	}
	err = enc.Encode(cats)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetCategory gets a Category from the database
// @summary gets a Category
// @router /categories/{id} [get]
// @tags categories
// @id getCategory
// @Param id path integer true "Category ID"
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 404 {object} model.Errors "Not found"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetCategory(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	category, err := app.categoryPersister.SelectOneCategory(req.URL.String())
	if err == persist.ErrNoRows {
		NotFound(w, req)
		return
	} else if err != nil {
		serverError(w, err)
		return
	}
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PostCategory adds a new Category to the database
// @summary adds a new Category
// @router /categories [post]
// @tags categories
// @id addCategory
// @Param category body model.CategoryIn true "Add Category"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Category "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PostCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		OtherErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	if mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		OtherErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	in := model.CategoryIn{}
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		OtherErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	err = app.validate.Struct(in)
	if err != nil {
		ValidationErrorResponse(w, 400, err)
		return
	}
	category, err := app.categoryPersister.InsertCategory(in)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PatchCategory updates a Category in the database
// @summary updates a Category
// @router /categories/{id} [patch]
// @tags categories
// @id updateCategory
// @Param id path integer true "Category ID"
// @Param category body model.CategoryIn true "Update Category"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PatchCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		OtherErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}

	var in model.CategoryIn
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		OtherErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	err = app.validate.Struct(in)
	if err != nil {
		ValidationErrorResponse(w, 400, err)
		return
	}
	category, err := app.categoryPersister.UpdateCategory(req.URL.String(), in)
	if err == persist.ErrNoRows {
		// Not allowed to add a Category with PATCH
		NotFound(w, req)
		return
	} else if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeleteCategory deletes a Category from the database
// @summary deletes a Category
// @router /categories/{id} [delete]
// @tags categories
// @id deleteCategory
// @Param id path integer true "Category ID"
// @success 204 "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) DeleteCategory(w http.ResponseWriter, req *http.Request) {
	err := app.categoryPersister.DeleteCategory(req.URL.String())
	if err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
