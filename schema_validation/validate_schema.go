// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package schema_validation

import (
    "encoding/json"
    "github.com/pb33f/libopenapi-validator/errors"
    "github.com/pb33f/libopenapi-validator/helpers"
    "github.com/pb33f/libopenapi/datamodel/high/base"
    "github.com/pb33f/libopenapi/utils"
    "github.com/santhosh-tekuri/jsonschema/v5"
    "gopkg.in/yaml.v3"
    "strings"
)

func ValidateSchema(schema *base.Schema, payload []byte) (bool, []*errors.ValidationError) {

    var validationErrors []*errors.ValidationError

    // render the schema, to be used for validation
    renderedSchema, _ := schema.RenderInline()
    jsonSchema, _ := utils.ConvertYAMLtoJSON(renderedSchema)

    var decodedObj interface{}
    _ = json.Unmarshal(payload, &decodedObj)

    compiler := jsonschema.NewCompiler()
    _ = compiler.AddResource("schema.json", strings.NewReader(string(jsonSchema)))
    jsch, _ := compiler.Compile("schema.json")

    // 4. validate the object against the schema
    scErrs := jsch.Validate(decodedObj)
    if scErrs != nil {
        jk := scErrs.(*jsonschema.ValidationError)

        // flatten the validationErrors
        schFlatErrs := jk.BasicOutput().Errors
        var schemaValidationErrors []*errors.SchemaValidationFailure
        for q := range schFlatErrs {
            er := schFlatErrs[q]
            if er.KeywordLocation == "" || strings.HasPrefix(er.Error, "doesn't validate with") {
                continue // ignore this error, it's useless tbh, utter noise.
            }
            if er.Error != "" {

                // re-encode the schema.
                var renderedNode yaml.Node
                _ = yaml.Unmarshal(renderedSchema, &renderedNode)

                // locate the violated property in the schema
                located := LocateSchemaPropertyNodeByJSONPath(renderedNode.Content[0], er.KeywordLocation)
                violation := &errors.SchemaValidationFailure{
                    Reason:        er.Error,
                    Location:      er.KeywordLocation,
                    OriginalError: jk,
                }
                // if we have a location within the schema, add it to the error
                if located != nil {
                    // location of the violation within the rendered schema.
                    violation.Line = located.Line
                    violation.Column = located.Column
                }
                schemaValidationErrors = append(schemaValidationErrors, violation)
            }
        }

        // add the error to the list
        validationErrors = append(validationErrors, &errors.ValidationError{
            ValidationType:         helpers.Schema,
            Message:                "schema does not pass validation",
            Reason:                 "Schema failed to validated against the contract requirements",
            SpecLine:               schema.GoLow().Type.KeyNode.Line,
            SpecCol:                schema.GoLow().Type.KeyNode.Column,
            SchemaValidationErrors: schemaValidationErrors,
            HowToFix:               errors.HowToFixInvalidSchema,
            Context:                string(renderedSchema), // attach the rendered schema to the error
        })
    }
    if len(validationErrors) > 0 {
        return false, validationErrors
    }
    return true, nil
}