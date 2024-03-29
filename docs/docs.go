// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/dl/{fileId}": {
            "get": {
                "produces": [
                    "application/octet-stream",
                    "application/json"
                ],
                "tags": [
                    "File controller"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "File id",
                        "name": "fileId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    }
                }
            }
        },
        "/kinozal/rss": {
            "get": {
                "produces": [
                    "text/xml",
                    "application/json"
                ],
                "tags": [
                    "Kinozal controller"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Rss"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    }
                }
            }
        },
        "/lostfilm/rss": {
            "get": {
                "produces": [
                    "text/xml",
                    "application/json"
                ],
                "tags": [
                    "LostFilm controller"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Quality filter",
                        "name": "quality",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Rss"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    }
                }
            }
        },
        "/proxy": {
            "get": {
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Proxy controller"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Url for proxied GET request",
                        "name": "url",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "override content-type header",
                        "name": "responseHeaderContentType",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        },
        "/twitch/messages": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Twitch controller"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Channel filter",
                        "name": "channel",
                        "in": "query"
                    },
                    {
                        "maximum": 100,
                        "type": "integer",
                        "description": "Message list limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/twitch.ChatMessage"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    }
                }
            }
        },
        "/twitch/tushqa": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Twitch controller"
                ],
                "parameters": [
                    {
                        "maximum": 100,
                        "type": "integer",
                        "description": "Quotes limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/twitch.TushqaQuote"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/web.HTTPError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "twitch.ChatMessage": {
            "type": "object",
            "properties": {
                "channel": {
                    "type": "string"
                },
                "created": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "originalTime": {
                    "type": "string"
                },
                "raw": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/twitch.ChatUser"
                }
            }
        },
        "twitch.ChatUser": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "twitch.TushqaQuote": {
            "type": "object",
            "properties": {
                "channel": {
                    "type": "string"
                },
                "created": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "web.HTTPError": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer",
                    "example": 400
                },
                "message": {
                    "type": "string",
                    "example": "status bad request"
                }
            }
        },
        "web.Rss": {
            "type": "object",
            "properties": {
                "channel": {
                    "$ref": "#/definitions/web.RssChannel"
                },
                "version": {
                    "type": "string"
                },
                "xmlname": {
                    "type": "object"
                }
            }
        },
        "web.RssChannel": {
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.RssChannelItem"
                    }
                },
                "lastBuildDate": {
                    "type": "string"
                },
                "link": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "web.RssChannelItem": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "link": {
                    "type": "string"
                },
                "originalDate": {
                    "type": "string"
                },
                "originalUrl": {
                    "type": "string"
                },
                "pubDate": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                },
                "uid": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
