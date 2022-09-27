package main

import (
	"fmt"
	"github.com/pseudomuto/protokit"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"math"
)

// CustomOptions stores all custom options defined by proto files being compiled
type CustomOptions struct {
	FileOptions      map[int32]*CustomOptionDef `json:"file_options"`
	MessageOptions   map[int32]*CustomOptionDef `json:"message_options"`
	FieldOptions     map[int32]*CustomOptionDef `json:"field_options"`
	EnumOptions      map[int32]*CustomOptionDef `json:"enum_options"`
	EnumValueOptions map[int32]*CustomOptionDef `json:"enum_value_options"`
	ServiceOptions   map[int32]*CustomOptionDef `json:"service_options"`
	MethodOptions    map[int32]*CustomOptionDef `json:"method_options"`
}

// CustomOptionDef is a custom option defined by a proto file
type CustomOptionDef struct {
	Index    int32
	Name     string
	FullName string
	Type     descriptorpb.FieldDescriptorProto_Type
	TypeName string
}

func NewCustomOptions() *CustomOptions {
	return &CustomOptions{
		FileOptions:      make(map[int32]*CustomOptionDef),
		MessageOptions:   make(map[int32]*CustomOptionDef),
		FieldOptions:     make(map[int32]*CustomOptionDef),
		EnumOptions:      make(map[int32]*CustomOptionDef),
		EnumValueOptions: make(map[int32]*CustomOptionDef),
		ServiceOptions:   make(map[int32]*CustomOptionDef),
		MethodOptions:    make(map[int32]*CustomOptionDef),
	}
}

// parseAllCustomOptionDefinitions finds all custom options defined by any of `files` and returns
// a struct containing them
func parseAllCustomOptionDefinitions(files []*protokit.FileDescriptor) *CustomOptions {
	ret := NewCustomOptions()

	//Loop through all extensions defined by all files
	for _, file := range files {
		for _, ext := range file.GetExtensions() {
			// To add support for message, file, etc. options, add a case statement here
			switch ext.GetExtendee() {
			case ".google.protobuf.FileOptions":
				ret.FileOptions[ext.GetNumber()] = parseCustomOption(ext)
			case ".google.protobuf.MessageOptions":
				ret.MessageOptions[ext.GetNumber()] = parseCustomOption(ext)
			case ".google.protobuf.FieldOptions":
				ret.FieldOptions[ext.GetNumber()] = parseCustomOption(ext)
			case ".google.protobuf.EnumOptions":
				ret.EnumOptions[ext.GetNumber()] = parseCustomOption(ext)
			case ".google.protobuf.EnumValueOptions":
				ret.EnumValueOptions[ext.GetNumber()] = parseCustomOption(ext)
			case ".google.protobuf.ServiceOptions":
				ret.ServiceOptions[ext.GetNumber()] = parseCustomOption(ext)
			case ".google.protobuf.MethodOptions":
				ret.MethodOptions[ext.GetNumber()] = parseCustomOption(ext)
			}
		}
	}

	return ret
}

// parseAllCustomOptionValues parses all the values of the custom options set on files/messages/fields/services/etc.
// and updates `context` with the parsed values
func parseAllCustomOptionValues(context *Context) {
	for _, file := range context.Files {
		file.Options = parseFileOptions(file.Descriptor, context)
	}
	for _, message := range context.Messages {
		message.Options = parseMessageOptions(message.Descriptor, context)
	}
	for _, field := range context.Fields {
		field.Options = parseFieldOptions(field.Descriptor, context)
	}
	for _, enum := range context.Enums {
		enum.Options = parseEnumOptions(enum.Descriptor, context)
	}
	for _, enumVal := range context.EnumValues {
		enumVal.Options = parseEnumValueOptions(enumVal.Descriptor, context)
	}
	for _, service := range context.Services {
		service.Options = parseServiceOptions(service.Descriptor, context)
	}
	for _, method := range context.Methods {
		method.Options = parseMethodOptions(method.Descriptor, context)
	}
}

// parseCustomOption parses a custom option
func parseCustomOption(ext *protokit.ExtensionDescriptor) *CustomOptionDef {
	ret := &CustomOptionDef{Index: ext.GetNumber(), Name: ext.GetName(), Type: ext.GetType(), TypeName: ext.GetTypeName()}

	// Lil' hack to get the FQN of the option
	ret.FullName = ext.GetPackage() + "." + ext.GetName()

	// `ext.GetTypeName()` will return an empty string if `Type` is a Message
	if len(ret.TypeName) == 0 {
		ret.TypeName = ret.Type.String()
	}

	return ret
}

func parseRawOptions(entityName string, raw protoreflect.RawFields, optionsDB *map[int32]*CustomOptionDef) map[string]interface{} {
	ret := make(map[string]interface{})

	size := len(raw)
	consumed := 0

	if size <= 0 {
		return nil
	}

	// Handle each option
	for consumed < size {
		// Read tag
		optionIndex, wireType, length := protowire.ConsumeTag(raw[consumed:])
		if length < 0 {
			fmt.Printf("FAILED PARSING OPTIONS FOR ENTITY %s: %v", entityName, raw)
			return nil
		}
		consumed += length

		// Find option
		optionDef, found := (*optionsDB)[int32(optionIndex)]

		if !found {
			continue
		}

		var opt interface{}

		uintVal := uint64(0)
		bytesVal := make([]byte, 0)
		valLen := 0

		// First, we need to decode value based on its WIRE type
		switch wireType {
		case protowire.VarintType:
			uintVal, valLen = protowire.ConsumeVarint(raw[consumed:])
		case protowire.Fixed32Type:
			val, i32ValLen := protowire.ConsumeFixed32(raw[consumed:])
			uintVal = uint64(val)
			valLen = i32ValLen
		case protowire.Fixed64Type:
			uintVal, valLen = protowire.ConsumeFixed64(raw[consumed:])
		case protowire.BytesType:
			bytesVal, valLen = protowire.ConsumeBytes(raw[consumed:])
		case protowire.StartGroupType:
			bytesVal, valLen = protowire.ConsumeGroup(optionIndex, raw[consumed:])
		case protowire.EndGroupType:
			// Should not get here
		default:
		}

		consumed += valLen

		//Next, we need to treat the decoded value differently based on the type it's declared as in the proto file.
		//This is because, for example, `sint32` and `int32` are both encoded as varints, but `sint32` is zigzag-encoded.
		switch optionDef.Type {
		case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
			opt = math.Float64frombits(uintVal)
		case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
			// TODO: proper float parsing
			opt = math.Float32frombits(uint32(uintVal))
		case descriptorpb.FieldDescriptorProto_TYPE_INT64,
			descriptorpb.FieldDescriptorProto_TYPE_INT32,
			descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
			descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
			// TODO: proper int parsing
			opt = int64(uintVal)
		case descriptorpb.FieldDescriptorProto_TYPE_UINT64,
			descriptorpb.FieldDescriptorProto_TYPE_UINT32,
			descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
			descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
			opt = uintVal
		case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
			// No ternary expression in Go? Gross!
			if uintVal == 0 {
				opt = false
			} else {
				opt = true
			}
		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
			opt = string(bytesVal)
		case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
			opt = "TODO: PARSE GROUPS"
		case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			// TODO: PARSE MESSAGES
			opt = optionDef.TypeName
		case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
			opt = bytesVal
		case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
			opt = struct {
				EnumType string `json:"enum_type"`
				EnumVal  uint64 `json:"enum_value"`
			}{
				StripStartingPeriod(optionDef.TypeName),
				uintVal,
			}
		case descriptorpb.FieldDescriptorProto_TYPE_SINT32,
			descriptorpb.FieldDescriptorProto_TYPE_SINT64:
			opt = protowire.DecodeZigZag(uintVal)
		}

		ret[optionDef.FullName] = opt
	}

	return ret
}

/**
 *
 * Here be the consequence of Golang's type system
 *
 * May you look upon this and despair
 *
 */

// parseFileOptions parses options on a file,
// mapping them to CustomOptions which were discovered during `parseAllCustomOptionDefinitions`
func parseFileOptions(file *protokit.FileDescriptor, context *Context) map[string]interface{} {
	ret := make(map[string]interface{})

	options := file.GetOptions()

	// Handle some common pre-defined (non-custom) options
	if options.GetDeprecated() {
		ret["deprecated"] = true
	}
	// End handling pre-defined options

	//`options` is a protobuf message -- custom options will be in its unknown fields
	raw := options.ProtoReflect().GetUnknown()

	// Parse the options from the raw bytes of the message and store them in ret
	for k, v := range parseRawOptions(file.GetName(), raw, &context.CustomOptions.FileOptions) {
		ret[k] = v
	}

	return ret
}

// parseMessageOptions parses options on a message,
// mapping them to CustomOptions which were discovered during `parseAllCustomOptionDefinitions`
func parseMessageOptions(message *protokit.Descriptor, context *Context) map[string]interface{} {
	ret := make(map[string]interface{})

	options := message.GetOptions()

	// Handle some common pre-defined (non-custom) options
	if options.GetDeprecated() {
		ret["deprecated"] = true
	}
	// End handling pre-defined options

	//`options` is a protobuf message -- custom options will be in its unknown fields
	raw := options.ProtoReflect().GetUnknown()

	// Parse the options from the raw bytes of the message and store them in ret
	for k, v := range parseRawOptions(message.GetFullName(), raw, &context.CustomOptions.MessageOptions) {
		ret[k] = v
	}

	return ret
}

// parseFieldOptions parses options on a field,
// mapping them to CustomOptions which were discovered during `parseAllCustomOptionDefinitions`
func parseFieldOptions(field *protokit.FieldDescriptor, context *Context) map[string]interface{} {
	ret := make(map[string]interface{})

	options := field.GetOptions()

	// Handle some common pre-defined (non-custom) options
	if options.GetDeprecated() {
		ret["deprecated"] = true
	}
	// End handling pre-defined options

	//`options` is a protobuf message -- custom options will be in its unknown fields
	raw := options.ProtoReflect().GetUnknown()

	// Parse the options from the raw bytes of the message and store them in ret
	for k, v := range parseRawOptions(field.GetFullName(), raw, &context.CustomOptions.FieldOptions) {
		ret[k] = v
	}

	return ret
}

// parseEnumOptions parses options on an enum,
// mapping them to CustomOptions which were discovered during `parseAllCustomOptionDefinitions`
func parseEnumOptions(enum *protokit.EnumDescriptor, context *Context) map[string]interface{} {
	ret := make(map[string]interface{})

	options := enum.GetOptions()

	// Handle some common pre-defined (non-custom) options
	if options.GetDeprecated() {
		ret["deprecated"] = true
	}
	// End handling pre-defined options

	//`options` is a protobuf message -- custom options will be in its unknown fields
	raw := options.ProtoReflect().GetUnknown()

	// Parse the options from the raw bytes of the message and store them in ret
	for k, v := range parseRawOptions(enum.GetFullName(), raw, &context.CustomOptions.EnumOptions) {
		ret[k] = v
	}

	return ret
}

// parseEnumValueOptions parses options on an enum value,
// mapping them to CustomOptions which were discovered during `parseAllCustomOptionDefinitions`
func parseEnumValueOptions(enumVal *protokit.EnumValueDescriptor, context *Context) map[string]interface{} {
	ret := make(map[string]interface{})

	options := enumVal.GetOptions()

	// Handle some common pre-defined (non-custom) options
	if options.GetDeprecated() {
		ret["deprecated"] = true
	}
	// End handling pre-defined options

	//`options` is a protobuf message -- custom options will be in its unknown fields
	raw := options.ProtoReflect().GetUnknown()

	// Parse the options from the raw bytes of the message and store them in ret
	for k, v := range parseRawOptions(enumVal.GetFullName(), raw, &context.CustomOptions.EnumValueOptions) {
		ret[k] = v
	}

	return ret
}

// parseServiceOptions parses options on a service,
// mapping them to CustomOptions which were discovered during `parseAllCustomOptionDefinitions`
func parseServiceOptions(service *protokit.ServiceDescriptor, context *Context) map[string]interface{} {
	ret := make(map[string]interface{})

	options := service.GetOptions()

	// Handle some common pre-defined (non-custom) options
	if options.GetDeprecated() {
		ret["deprecated"] = true
	}
	// End handling pre-defined options

	//`options` is a protobuf message -- custom options will be in its unknown fields
	raw := options.ProtoReflect().GetUnknown()

	// Parse the options from the raw bytes of the message and store them in ret
	for k, v := range parseRawOptions(service.GetFullName(), raw, &context.CustomOptions.ServiceOptions) {
		ret[k] = v
	}

	return ret
}

// parseMethodOptions parses options on a method,
// mapping them to CustomOptions which were discovered during `parseAllCustomOptionDefinitions`
func parseMethodOptions(method *protokit.MethodDescriptor, context *Context) map[string]interface{} {
	ret := make(map[string]interface{})

	options := method.GetOptions()

	// Handle some common pre-defined (non-custom) options
	if options.GetDeprecated() {
		ret["deprecated"] = true
	}
	// End handling pre-defined options

	//`options` is a protobuf message -- custom options will be in its unknown fields
	raw := options.ProtoReflect().GetUnknown()

	// Parse the options from the raw bytes of the message and store them in ret
	for k, v := range parseRawOptions(method.GetFullName(), raw, &context.CustomOptions.MethodOptions) {
		ret[k] = v
	}

	return ret
}
