package main

import (
	"fmt"
	"github.com/pseudomuto/protokit"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/types/descriptorpb"
	"math"
)

// CustomOptions stores all custom options defined by proto files being compiled
type CustomOptions struct {
	FieldOptions map[int32]*CustomOptionDef `json:"field_options"`
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
		FieldOptions: make(map[int32]*CustomOptionDef),
	}
}

// parseAllOptions finds all custom options defined by any of `files` and returns
// a struct containing them
func parseAllOptions(files []*protokit.FileDescriptor) *CustomOptions {
	ret := &CustomOptions{FieldOptions: make(map[int32]*CustomOptionDef)}

	//Loop through all extensions defined by all files
	for _, file := range files {
		for _, ext := range file.GetExtensions() {
			// To add support for message, file, etc. options, add a case statement here
			switch ext.GetExtendee() {
			case ".google.protobuf.FieldOptions":
				ret.FieldOptions[ext.GetNumber()] = parseCustomOption(ext)
			}
		}
	}

	return ret
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

// parseFieldOptions parses options on a field,
// mapping them to CustomOptions which were discovered during `parseAllOptions`
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
			fmt.Printf("FAILED PARSING OPTIONS FOR FIELD %s: %v", field.GetFullName(), raw)
			return nil
		}
		consumed += length

		// Find option
		optionDef, found := context.CustomOptions.FieldOptions[int32(optionIndex)]

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
			opt = "TODO: PARSE ENUM"
		case descriptorpb.FieldDescriptorProto_TYPE_SINT32,
			descriptorpb.FieldDescriptorProto_TYPE_SINT64:
			opt = protowire.DecodeZigZag(uintVal)
		}

		ret[optionDef.FullName] = opt
	}

	return ret
}
