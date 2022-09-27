package main

import (
	"bytes"
	"encoding/json"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pseudomuto/protokit"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"log"
	"strings"
)

// Required struct for protokit -- its `Generate()` method will be called.
type plugin struct{}

// Entry point of plugin -- takes input from `protoc` and handles it
func main() {
	// Run plugin via `protokit`
	if err := protokit.RunPlugin(new(plugin)); err != nil {
		log.Fatal(err)
	}
}

// Generate generates output from incoming proto files
func (p *plugin) Generate(req *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	// Prepare context
	context := NewContext()

	// Parse request via protokit
	descriptors := protokit.ParseCodeGenRequest(req)

	// First parse out all custom options defined in all files
	context.CustomOptions = parseAllCustomOptionDefinitions(descriptors)

	// Then, parse everything defined in each file, EXCEPT for the custom options defined on resources
	for _, d := range descriptors {
		parseFile(d, context)
	}

	// Finally, parse all the custom options that were set on fields/etc.
	// We have to do this step last because a custom option might be of a type that isn't defined until everything
	// has been parsed
	parseAllCustomOptionValues(context)

	// Encode to JSON
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")
	encodeResult := encoder.Encode(context)

	if encodeResult != nil {
		return nil, encodeResult
	}

	// Determine the output filename
	filename := "output.json"
	if req.GetParameter() != "" {
		filename = req.GetParameter()
	}

	// Tell protoc we're done
	ret := new(plugin_go.CodeGeneratorResponse)
	ret.File = append(ret.File, &plugin_go.CodeGeneratorResponse_File{
		Name:    proto.String(filename),
		Content: proto.String(buf.String()),
	})

	// Tell `protoc` that we support optional proto3 fields
	// We need to heap-allocate this so we can do pointer stuff because the response object
	// has a pointer to a uint64 which might be nil (why??)
	ret.SupportedFeatures = new(uint64)
	*ret.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

	return ret, nil
}

// parseFile parses a protobuf file and all its constituent parts
func parseFile(fileProto *protokit.FileDescriptor, context *Context) {
	file := &File{
		Descriptor: fileProto,

		Name:        fileProto.GetName(),
		Package:     fileProto.GetPackage(),
		Description: fileProto.GetPackageComments().String(),

		Services:   make([]string, 0),
		Methods:    make([]string, 0),
		Messages:   make([]string, 0),
		Fields:     make([]string, 0),
		Enums:      make([]string, 0),
		EnumValues: make([]string, 0),
	}

	// Handle all services in fileProto
	for _, service := range fileProto.GetServices() {
		parseService(service, context, file)
	}

	// Handle all messages in fileProto
	for _, msg := range fileProto.GetMessages() {
		parseMessage(msg, context, file, nil)
	}

	// Handle all enums in fileProto
	for _, enum := range fileProto.GetEnums() {
		parseEnum(enum, context, file, nil)
	}

	// Store file in context
	context.StoreFile(file)

	// TODO: Handle extensions defined in a fileProto?
}

// parseMessage parses a protobuf message and its fields
func parseMessage(messageProto *protokit.Descriptor, context *Context, declFile *File, declMessage *Message) {
	message := &Message{
		Descriptor:  messageProto,
		Name:        messageProto.GetName(),
		FullName:    GetFQN(messageProto.GetFullName()),
		Description: messageProto.GetComments().String(),
		IsMapEntry:  messageProto.Options.GetMapEntry(),

		Fields:   make([]string, 0),
		Messages: make([]string, 0),
		Enums:    make([]string, 0),
	}

	//Put messageProto into declFile.Messages and, if non-null, declMessage.Messages
	declFile.Messages = append(declFile.Messages, message.FullName)
	if declMessage != nil {
		declMessage.Messages = append(declMessage.Messages, message.FullName)
	}

	// Handle all fields in messageProto
	for _, fd := range messageProto.GetMessageFields() {
		parseField(fd, context, declFile, message)
	}

	// Handle all sub-messages in messageProto
	for _, sm := range messageProto.GetMessages() {
		parseMessage(sm, context, declFile, message)
	}

	// Handle all sub-enums in messageProto
	for _, enum := range messageProto.GetEnums() {
		parseEnum(enum, context, declFile, message)
	}

	//TODO: Handle extensions defined in a messageProto?

	//Store message in context
	context.StoreMessage(message, messageProto)
}

// parseField parses a field in a protobuf message, and its options
func parseField(fieldProto *protokit.FieldDescriptor, context *Context, declFile *File, declMessage *Message) {
	// Figure out type names
	typeName := fieldProto.GetTypeName()
	fullTypeName := fieldProto.GetType().String()

	// If typeName is length 0, it's a proto primitive
	if len(typeName) == 0 {
		typeName = strings.Replace(fieldProto.GetType().String(), "TYPE_", "", -1)
		typeName = strings.ToLower(typeName)
		fullTypeName = typeName
	} else {
		// Otherwise, `typeName` is the fully-qualified type,
		// and `fullTypeName` will be `TYPE_MESSAGE`.

		fullTypeName = typeName
		split := strings.Split(typeName, ".")
		typeName = split[len(split)-1]
	}

	// Determine FQN
	fqn := GetFQN(fieldProto.GetFullName())

	field := &Field{
		Descriptor:  fieldProto,
		Name:        fieldProto.GetName(),
		FullName:    fqn,
		Label:       fieldProto.GetLabel().String(),
		Type:        GetFQN(typeName),
		FullType:    GetFQN(fullTypeName),
		Description: fieldProto.GetComments().String(),
	}

	// Store fieldProto in declFile and declMessage
	declFile.Fields = append(declFile.Fields, fqn)
	declMessage.Fields = append(declMessage.Fields, fqn)

	//Store field in context
	context.StoreField(field, fieldProto)
}

func parseEnum(enumProto *protokit.EnumDescriptor, context *Context, declFile *File, declMessage *Message) {
	enum := &Enum{
		Descriptor:  enumProto,
		Name:        enumProto.GetName(),
		FullName:    GetFQN(enumProto.GetFullName()),
		Description: enumProto.GetComments().String(),
		Values:      make([]string, 0),
	}

	//Store enum in declFile.Enums and, if non-null, declMessage.Enums
	declFile.Enums = append(declFile.Enums, enum.FullName)
	if declMessage != nil {
		declMessage.Enums = append(declMessage.Enums, enum.FullName)
	}

	for _, val := range enumProto.GetValues() {
		parseEnumValue(val, context, declFile, enum)
	}

	// Index enumProto
	context.StoreEnum(enum, enumProto)
}

func parseEnumValue(enumValProto *protokit.EnumValueDescriptor, context *Context, declFile *File, declEnum *Enum) {
	enumVal := &EnumValue{
		Descriptor:  enumValProto,
		Name:        enumValProto.GetName(),
		FullName:    GetFQN(enumValProto.GetFullName()),
		Description: enumValProto.GetComments().String(),
		Value:       enumValProto.GetNumber(),
	}

	//Store enumVal in declFile.EnumValues and declEenum.Values
	declFile.EnumValues = append(declFile.EnumValues, enumVal.FullName)
	declEnum.Values = append(declEnum.Values, enumVal.FullName)

	//Store enumVal in context
	context.StoreEnumValue(enumVal, enumValProto)
}

// parseService parses a service in a protobuf file, and its methods
func parseService(serviceProto *protokit.ServiceDescriptor, context *Context, declFile *File) {
	service := &Service{
		Descriptor:  serviceProto,
		Name:        serviceProto.GetName(),
		FullName:    GetFQN(serviceProto.GetFullName()),
		Description: serviceProto.GetComments().String(),
		Methods:     make([]string, 0),
	}

	// Store service in declFile.Services
	declFile.Services = append(declFile.Services, service.FullName)

	for _, md := range serviceProto.GetMethods() {
		parseMethod(md, context, declFile, service)
	}

	//Store service in context
	context.StoreService(service, serviceProto)
}

// parseMethod parses a method in a service
func parseMethod(methodProto *protokit.MethodDescriptor, context *Context, declFile *File, declService *Service) {
	method := &Method{
		Descriptor:  methodProto,
		Name:        methodProto.GetName(),
		FullName:    GetFQN(methodProto.GetFullName()),
		InputType:   GetFQN(methodProto.GetInputType()),
		OutputType:  GetFQN(methodProto.GetOutputType()),
		Description: methodProto.GetComments().String(),
	}

	//Store method in declFile.Methods and declService.Methods
	declFile.Methods = append(declFile.Methods, method.FullName)
	declService.Methods = append(declService.Methods, method.FullName)

	// Store index in context
	context.StoreMethod(method, methodProto)
}
