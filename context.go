package main

import (
	"github.com/pseudomuto/protokit"
)

// Context carries context throughout the compilation process, and is output as JSON
type Context struct {
	CustomOptions *CustomOptions         `json:"-"`
	Index         map[string]*IndexEntry `json:"index"`

	Files      map[string]*File      `json:"files"`
	Services   map[string]*Service   `json:"services"`
	Methods    map[string]*Method    `json:"methods"`
	Messages   map[string]*Message   `json:"messages"`
	Fields     map[string]*Field     `json:"fields"`
	Enums      map[string]*Enum      `json:"enums"`
	EnumValues map[string]*EnumValue `json:"enum_values"`
}

type IndexEntry struct {
	Type       string `json:"type"`
	Collection string `json:"collection"`
	File       string `json:"file"`
	Parent     string `json:"parent"`
}

// File is a parsed protobuf file
type File struct {
	Descriptor *protokit.FileDescriptor `json:"-"`

	Name        string                 `json:"name"`
	Package     string                 `json:"package"`
	Description string                 `json:"description"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Services    []string               `json:"services"`
	Methods     []string               `json:"methods"`
	Messages    []string               `json:"messages"`
	Fields      []string               `json:"fields"`
	Enums       []string               `json:"enums"`
	EnumValues  []string               `json:"enum_values"`
}

// Service is a parsed service defined in a File
type Service struct {
	Descriptor *protokit.ServiceDescriptor `json:"-"`

	Name        string                 `json:"name"`
	FullName    string                 `json:"full_name"`
	Description string                 `json:"description"`
	Methods     []string               `json:"methods"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// Method is a parsed service method
type Method struct {
	Descriptor *protokit.MethodDescriptor `json:"-"`

	Name        string                 `json:"name"`
	FullName    string                 `json:"full_name"`
	InputType   string                 `json:"input_type"`
	OutputType  string                 `json:"output_type"`
	Description string                 `json:"description"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// Message is a parsed message defined in a file
type Message struct {
	Descriptor *protokit.Descriptor `json:"-"`

	Name        string                 `json:"name"`
	FullName    string                 `json:"full_name"`
	Description string                 `json:"description"`
	IsMapEntry  bool                   `json:"is_map_entry,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Fields      []string               `json:"fields"`
	Messages    []string               `json:"messages"`
	Enums       []string               `json:"enums"`
}

// Field is a parsed field defined in a Message
type Field struct {
	Descriptor *protokit.FieldDescriptor `json:"-"`

	Name        string                 `json:"name"`
	FullName    string                 `json:"full_name"`
	Label       string                 `json:"label"`
	Type        string                 `json:"type"`
	FullType    string                 `json:"full_type"`
	Description string                 `json:"description"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

type Enum struct {
	Descriptor *protokit.EnumDescriptor `json:"-"`

	Name        string                 `json:"name"`
	FullName    string                 `json:"full_name"`
	Description string                 `json:"description"`
	Values      []string               `json:"values"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

type EnumValue struct {
	Descriptor *protokit.EnumValueDescriptor `json:"-"`

	Name        string                 `json:"name"`
	FullName    string                 `json:"full_name"`
	Description string                 `json:"description"`
	Value       int32                  `json:"value"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

func NewContext() *Context {
	return &Context{
		CustomOptions: NewCustomOptions(),
		Index:         make(map[string]*IndexEntry),

		Files:      make(map[string]*File),
		Services:   make(map[string]*Service),
		Methods:    make(map[string]*Method),
		Messages:   make(map[string]*Message),
		Fields:     make(map[string]*Field),
		Enums:      make(map[string]*Enum),
		EnumValues: make(map[string]*EnumValue),
	}
}

// StoreFile stores a file in a context
func (ctx *Context) StoreFile(file *File) {
	ctx.Files[file.Name] = file
}

// StoreMessage stores a message in a context, and indexes it
func (ctx *Context) StoreMessage(message *Message, msgProto *protokit.Descriptor) {
	ctx.Messages[message.FullName] = message

	entry := &IndexEntry{
		Type:       "message",
		Collection: "messages",
		File:       msgProto.GetFile().GetName(),
	}

	if msgProto.GetParent() != nil {
		entry.Parent = GetFQN(msgProto.GetParent().GetFullName())
	}

	ctx.Index[GetFQN(msgProto.GetFullName())] = entry
}

// StoreField stores a field in a context, and indexes it
func (ctx *Context) StoreField(field *Field, fieldProto *protokit.FieldDescriptor) {
	ctx.Fields[field.FullName] = field

	entry := &IndexEntry{
		Type:       "field",
		Collection: "fields",
		File:       fieldProto.GetFile().GetName(),
	}

	if fieldProto.GetMessage() != nil {
		entry.Parent = GetFQN(fieldProto.GetMessage().GetFullName())
	}

	ctx.Index[GetFQN(fieldProto.GetFullName())] = entry
}

// StoreEnum stores an enum in a context, and indexes it
func (ctx *Context) StoreEnum(enum *Enum, enumProto *protokit.EnumDescriptor) {
	ctx.Enums[enum.FullName] = enum

	entry := &IndexEntry{
		Type:       "enum",
		Collection: "enums",
		File:       enumProto.GetFile().GetName(),
	}

	if enumProto.GetParent() != nil {
		entry.Parent = GetFQN(enumProto.GetParent().GetFullName())
	}

	ctx.Index[GetFQN(enumProto.GetFullName())] = entry
}

// StoreEnumValue stores an enum value in a context, and indexes it
func (ctx *Context) StoreEnumValue(enumValue *EnumValue, enumValueProto *protokit.EnumValueDescriptor) {
	ctx.EnumValues[enumValue.FullName] = enumValue

	entry := &IndexEntry{
		Type:       "enum_value",
		Collection: "enum_values",
		File:       enumValueProto.GetFile().GetName(),
	}

	if enumValueProto.GetEnum() != nil {
		entry.Parent = GetFQN(enumValueProto.GetEnum().GetFullName())
	}

	ctx.Index[GetFQN(enumValueProto.GetFullName())] = entry
}

// StoreService stores a service in a context, and indexes it
func (ctx *Context) StoreService(service *Service, serviceProto *protokit.ServiceDescriptor) {
	ctx.Services[service.FullName] = service

	entry := &IndexEntry{
		Type:       "service",
		Collection: "services",
		File:       serviceProto.GetFile().GetName(),
	}

	ctx.Index[GetFQN(serviceProto.GetFullName())] = entry
}

// StoreMethod stores a method in a context, and indexes it
func (ctx *Context) StoreMethod(method *Method, methodProto *protokit.MethodDescriptor) {
	ctx.Methods[method.FullName] = method

	entry := &IndexEntry{
		Type:       "method",
		Collection: "methods",
		File:       methodProto.GetFile().GetName(),
	}

	ctx.Index[GetFQN(methodProto.GetFullName())] = entry
}
