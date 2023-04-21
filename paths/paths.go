// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package paths

import (
	"fmt"
	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// FindPath will find the path in the document that matches the request path. If a successful match was found, then
// the first return value will be a pointer to the PathItem. The second return value will contain any validation errors
// that were picked up when locating the path. Number/Integer validation is performed in any path parameters in the request.
// The third return value will be the path that was found in the document, as it pertains to the contract, so all path
// parameters will not have been replaced with their values from the request - allowing model lookups.
func FindPath(request *http.Request, document *v3.Document) (*v3.PathItem, []*errors.ValidationError, string) {

	var validationErrors []*errors.ValidationError

	reqPathSegments := strings.Split(request.URL.Path, "/")
	if reqPathSegments[0] == "" {
		reqPathSegments = reqPathSegments[1:]
	}
	var pItem *v3.PathItem
	var foundPath string
pathFound:
	for path, pathItem := range document.Paths.PathItems {
		segs := strings.Split(path, "/")
		if segs[0] == "" {
			segs = segs[1:]
		}

		// collect path level params
		params := pathItem.Parameters

		switch request.Method {
		case http.MethodGet:
			if pathItem.Get != nil {
				p := append(params, pathItem.Get.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, errs := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					validationErrors = errs
					break pathFound
				}
			}
		case http.MethodPost:
			if pathItem.Post != nil {
				p := append(params, pathItem.Post.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, _ := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
			}
		case http.MethodPut:
			if pathItem.Put != nil {
				p := append(params, pathItem.Put.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, errs := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					validationErrors = errs
					break pathFound
				}
			}
		case http.MethodDelete:
			if pathItem.Delete != nil {
				p := append(params, pathItem.Delete.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, errs := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					validationErrors = errs
					break pathFound
				}
			}
		case http.MethodOptions:
			if pathItem.Options != nil {
				p := append(params, pathItem.Options.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, errs := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					validationErrors = errs
					break pathFound
				}
			}
		case http.MethodHead:
			if pathItem.Head != nil {
				p := append(params, pathItem.Head.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, errs := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					validationErrors = errs
					break pathFound
				}
			}
		case http.MethodPatch:
			if pathItem.Patch != nil {
				p := append(params, pathItem.Patch.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, errs := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					validationErrors = errs
					break pathFound
				}
			}
		case http.MethodTrace:
			if pathItem.Trace != nil {
				p := append(params, pathItem.Trace.Parameters...)
				// check for a literal match
				if request.URL.Path == path {
					pItem = pathItem
					foundPath = path
					break pathFound
				}
				if ok, errs := comparePaths(segs, reqPathSegments, p, request.URL.Path); ok {
					pItem = pathItem
					foundPath = path
					validationErrors = errs
					break pathFound
				}
			}
		}
	}
	if pItem == nil {
		validationErrors = append(validationErrors, &errors.ValidationError{
			ValidationType:    helpers.ParameterValidationPath,
			ValidationSubType: "missing",
			Message:           fmt.Sprintf("Path '%s' not found", request.URL.Path),
			Reason: fmt.Sprintf("The request contains a path of '%s' "+
				"however that path does not exist in the specification", request.URL.Path),
			SpecLine: -1,
			SpecCol:  -1,
		})
		return pItem, validationErrors, foundPath
	} else {
		return pItem, validationErrors, foundPath
	}
}

func comparePaths(mapped, requested []string,
	params []*v3.Parameter, path string) (bool, []*errors.ValidationError) {

	// check lengths first
	var pathErrors []*errors.ValidationError

	if len(mapped) != len(requested) {
		return false, nil // short circuit out
	}
	var imploded []string
	for i, seg := range mapped {
		s := seg
		// check for braces
		if strings.Contains(seg, "{") {
			s = requested[i]
		}
		// check param against type, check if it's a number or not, and if it validates.
		for p := range params {
			if params[p].In == helpers.Path {
				h := seg[1 : len(seg)-1]
				if params[p].Name == h {
					schema := params[p].Schema.Schema()
					for t := range schema.Type {

						switch schema.Type[t] {
						case helpers.String:
							// should not be a number.
							if _, err := strconv.ParseFloat(s, 64); err == nil {
								s = "&&FAIL&&"
							}
						case helpers.Number, helpers.Integer:
							// should not be a string.
							if _, err := strconv.ParseFloat(s, 64); err != nil {
								s = "&&FAIL&&"
							}
						}

						//if schema.Type[t] == helpers.Number || schema.Type[t] == helpers.Integer {
						//notaNumber := false
						// will return no error on floats or int

						//if notaNumber {
						//	pathErrors = append(pathErrors, &errors.ValidationError{
						//		ValidationType:    helpers.ParameterValidationPath,
						//		ValidationSubType: "number",
						//		Message: fmt.Sprintf("Match for path '%s', but the parameter "+
						//			"'%s' is not a number", path, s),
						//		Reason: fmt.Sprintf("The parameter '%s' is defined as a number, "+
						//			"but the value '%s' is not a number", h, s),
						//		SpecLine: params[p].GoLow().Schema.Value.Schema().Type.KeyNode.Line,
						//		SpecCol:  params[p].GoLow().Schema.Value.Schema().Type.KeyNode.Column,
						//		Context:  schema,
						//	})
						//}
						//}
					}
				}
			}
		}
		imploded = append(imploded, s)
	}
	l := filepath.Join(imploded...)
	r := filepath.Join(requested...)
	if l == r {
		return true, pathErrors
	}
	return false, pathErrors
}
