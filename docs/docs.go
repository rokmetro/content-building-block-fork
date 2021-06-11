// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/student_guides": {
            "get": {
                "security": [
                    {
                        "RokwireAuth AdminUserAuth": []
                    }
                ],
                "description": "Retrieves  all items",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "StudentGuides"
                ],
                "operationId": "getAllStudentGuides",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Retrieves  all items",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "StudentGuides"
                ],
                "operationId": "CreateStudentGuide",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/student_guides/{id}": {
            "get": {
                "security": [
                    {
                        "RokwireAuth AdminUserAuth": []
                    }
                ],
                "description": "Retrieves  all items",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "StudentGuides"
                ],
                "operationId": "GetStudentGuide",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Updates a student guide with the specified id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "StudentGuides"
                ],
                "operationId": "UpdateStudentGuide",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Deletes a student guide with the specified id",
                "tags": [
                    "StudentGuides"
                ],
                "operationId": "DeleteStudentGuide",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/version": {
            "get": {
                "description": "Gives the service version.",
                "produces": [
                    "text/plain"
                ],
                "operationId": "Version",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "AdminGroupAuth": {
            "type": "apiKey",
            "name": "GROUP",
            "in": "header"
        },
        "AdminUserAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header (add Bearer prefix to the Authorization value)"
        },
        "RokwireAuth": {
            "type": "apiKey",
            "name": "ROKWIRE-API-KEY",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "0.4.0",
	Host:        "localhost",
	BasePath:    "/content",
	Schemes:     []string{"https"},
	Title:       "Rokwire Content Building Block API",
	Description: "Rokwire Content Building Block API Documentation.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
