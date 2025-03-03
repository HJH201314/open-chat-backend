// Package docs Code generated by swaggo/swag. DO NOT EDIT
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
        "/chat/completion/stream/{session_id}": {
            "post": {
                "description": "流式输出聊天",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/event-stream"
                ],
                "tags": [
                    "Chat"
                ],
                "summary": "流式输出聊天",
                "parameters": [
                    {
                        "type": "string",
                        "description": "会话 ID",
                        "name": "session_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "用户输入及参数",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/chat.CompletionStream.userInput"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/chat/config/models": {
            "get": {
                "description": "获取所有模型",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "config"
                ],
                "summary": "获取所有模型",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-array_models_ModelCache"
                        }
                    }
                }
            }
        },
        "/chat/message/list/{session_id}": {
            "get": {
                "description": "获取消息",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Message"
                ],
                "summary": "获取消息",
                "parameters": [
                    {
                        "type": "string",
                        "description": "会话 ID",
                        "name": "session_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "page_num",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "page_size",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "name": "sort_expr",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "返回数据",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-chat_GetMessages_resType"
                        }
                    }
                }
            }
        },
        "/chat/session/del/{session_id}": {
            "post": {
                "description": "删除会话",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Session"
                ],
                "summary": "删除会话",
                "parameters": [
                    {
                        "type": "string",
                        "description": "会话 ID",
                        "name": "session_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-bool"
                        }
                    }
                }
            }
        },
        "/chat/session/new": {
            "post": {
                "description": "创建会话",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Session"
                ],
                "summary": "创建会话",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-string"
                        }
                    }
                }
            }
        },
        "/manage/provider/create": {
            "post": {
                "description": "创建 API 提供商",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Provider"
                ],
                "summary": "创建 API 提供商",
                "parameters": [
                    {
                        "description": "API 提供商参数",
                        "name": "provider",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.Provider"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "成功创建的 API 提供商",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-models_Provider"
                        }
                    }
                }
            }
        },
        "/manage/provider/delete/{provider_id}": {
            "post": {
                "description": "删除 API 提供商",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Provider"
                ],
                "summary": "删除 API 提供商",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "API 提供商 ID",
                        "name": "provider_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "删除成功与否",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-bool"
                        }
                    }
                }
            }
        },
        "/manage/provider/list": {
            "get": {
                "description": "批量获取 API 提供商",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Provider"
                ],
                "summary": "批量获取 API 提供商",
                "responses": {
                    "200": {
                        "description": "API 提供商列表",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-array_models_Provider"
                        }
                    }
                }
            }
        },
        "/manage/provider/update": {
            "post": {
                "description": "更新 API 提供商",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Provider"
                ],
                "summary": "更新 API 提供商",
                "parameters": [
                    {
                        "description": "API 提供商参数",
                        "name": "provider",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.Provider"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "更新成功与否",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-bool"
                        }
                    }
                }
            }
        },
        "/manage/provider/{provider_id}": {
            "get": {
                "description": "获取 API 提供商",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Provider"
                ],
                "summary": "获取 API 提供商",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "API 提供商 ID",
                        "name": "provider_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "API 提供商",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-models_Provider"
                        }
                    }
                }
            }
        },
        "/user/login": {
            "post": {
                "description": "用户登录",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "用户登录",
                "parameters": [
                    {
                        "description": "登录请求",
                        "name": "req",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.Login.loginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "login successfully",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-models_User"
                        }
                    }
                }
            }
        },
        "/user/ping": {
            "post": {
                "description": "检测客户端登录态",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "检测客户端登录态",
                "responses": {
                    "200": {
                        "description": "user is online",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-models_User"
                        }
                    },
                    "404": {
                        "description": "user not found",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-any"
                        }
                    }
                }
            }
        },
        "/user/refresh": {
            "get": {
                "description": "刷新登录态",
                "tags": [
                    "User"
                ],
                "summary": "刷新登录态",
                "parameters": [
                    {
                        "type": "string",
                        "description": "刷新用 Token",
                        "name": "X-Refresh-Token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "nothing",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/register": {
            "post": {
                "description": "用户注册",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "用户注册",
                "parameters": [
                    {
                        "description": "注册请求",
                        "name": "req",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.Register.registerRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "register successfully",
                        "schema": {
                            "$ref": "#/definitions/entities.CommonResponse-bool"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "chat.CompletionStream.userInput": {
            "type": "object",
            "required": [
                "model_name",
                "provider_name",
                "question"
            ],
            "properties": {
                "enable_context": {
                    "type": "boolean"
                },
                "model_name": {
                    "description": "Model.Name 准确的模型名称",
                    "type": "string"
                },
                "provider_name": {
                    "description": "Provider.Name 准确的供应商名称",
                    "type": "string"
                },
                "question": {
                    "type": "string"
                },
                "system_prompt": {
                    "description": "系统提示词",
                    "type": "string"
                }
            }
        },
        "chat.GetMessages.resType": {
            "type": "object",
            "properties": {
                "messages": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Message"
                    }
                },
                "next_page": {
                    "type": "integer"
                }
            }
        },
        "entities.CommonResponse-any": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据"
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "entities.CommonResponse-array_models_ModelCache": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.ModelCache"
                    }
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "entities.CommonResponse-array_models_Provider": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Provider"
                    }
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "entities.CommonResponse-bool": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据",
                    "type": "boolean"
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "entities.CommonResponse-chat_GetMessages_resType": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据",
                    "allOf": [
                        {
                            "$ref": "#/definitions/chat.GetMessages.resType"
                        }
                    ]
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "entities.CommonResponse-models_Provider": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.Provider"
                        }
                    ]
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "entities.CommonResponse-models_User": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.User"
                        }
                    ]
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "entities.CommonResponse-string": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "代码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据",
                    "type": "string"
                },
                "msg": {
                    "description": "消息",
                    "type": "string"
                }
            }
        },
        "models.APIKey": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "key": {
                    "description": "API 密钥",
                    "type": "string"
                },
                "provider_id": {
                    "description": "外键，指向 Provider",
                    "type": "integer"
                }
            }
        },
        "models.Message": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "model_id": {
                    "description": "回复所使用的模型",
                    "type": "integer"
                },
                "reasoning_content": {
                    "type": "string"
                },
                "role": {
                    "description": "user/assistant/system",
                    "type": "string"
                },
                "session_id": {
                    "type": "string"
                }
            }
        },
        "models.Model": {
            "type": "object",
            "properties": {
                "config": {
                    "description": "使用 JSON 储存配置",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.ModelConfig"
                        }
                    ]
                },
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "description": "额外模型描述",
                    "type": "string"
                },
                "display_name": {
                    "description": "对外展示模型名称",
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "description": "模型名称",
                    "type": "string"
                },
                "provider_id": {
                    "description": "关联的 Provider ID",
                    "type": "integer"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.ModelCache": {
            "type": "object",
            "properties": {
                "config": {
                    "description": "使用 JSON 储存配置",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.ModelConfig"
                        }
                    ]
                },
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "description": "额外模型描述",
                    "type": "string"
                },
                "display_name": {
                    "description": "对外展示模型名称",
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "description": "模型名称",
                    "type": "string"
                },
                "provider_display_name": {
                    "description": "关联的 Provider DisplayName",
                    "type": "string"
                },
                "provider_id": {
                    "description": "关联的 Provider ID",
                    "type": "integer"
                },
                "provider_name": {
                    "description": "关联的 Provider Name",
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.ModelConfig": {
            "type": "object",
            "properties": {
                "allow_system_prompt": {
                    "description": "是否允许用户自行修改系统提示",
                    "type": "boolean"
                },
                "default_temperature": {
                    "description": "默认温度",
                    "type": "number"
                },
                "frequency_penalty": {
                    "type": "number"
                },
                "max_tokens": {
                    "type": "integer"
                },
                "presence_penalty": {
                    "type": "number"
                },
                "system_prompt": {
                    "description": "预设系统提示",
                    "type": "string"
                },
                "top_p": {
                    "type": "number"
                }
            }
        },
        "models.Permission": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "description": "权限描述",
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "description": "权限名称",
                    "type": "string"
                },
                "path": {
                    "description": "权限路径（一般与名称相同）",
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.Provider": {
            "type": "object",
            "properties": {
                "api_keys": {
                    "description": "一对多关系，与 APIKey 模型关联",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.APIKey"
                    }
                },
                "base_url": {
                    "description": "API 的基本 URL",
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "description": "额外提供商描述",
                    "type": "string"
                },
                "display_name": {
                    "description": "对外展示提供商名称",
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "models": {
                    "description": "一对多关系，与 Model 模型关联",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Model"
                    }
                },
                "name": {
                    "description": "提供商名称",
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.Role": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "description": "角色描述",
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "description": "角色名称",
                    "type": "string"
                },
                "permissions": {
                    "description": "多对多关联",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Permission"
                    }
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "models.User": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "roles": {
                    "description": "用户与角色之间的多对多关系",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Role"
                    }
                },
                "updated_at": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.Login.loginRequest": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.Register.registerRequest": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
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
