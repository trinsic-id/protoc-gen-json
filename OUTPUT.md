# Output Format

## Overview

`protoc-gen-json` exports two main resources in its output file: 

- An ***index*** which maps the Fully-Qualfied Name (FQN) of any object (Message, Field, Enum, Enum Value, Service, Method, etc.) to an *index entry*, which details:
    - Type
    - Collection (_see below_)
    - File
      - Where the object was defined
    - Parent
        - Object within which this object was defined
        - EG, a Field's `parent` is its Message.
          - Enums and Messages could either have no parent (defined at top level of file), or their parent could be a Message.
- A set of ***collections***, one for each type of object (`files`, `messages`, `fields`, `services`, `methods`, `enums`, `enum_values`)
    - **Note**: `files` is the only collection which is _not_ indexed by `index`.


## The Index

The index is a flat map, representing _every single object_ defined in all compiled protobuf files. This means every message, field, enum, enum value, etc.

`index` can be used to resolve references across all boundaries. 

For example, a `message` has an array of `fields`, which is simply a list of strings: the fully-qualified names of the field objects which this Message defines.

For example, the following protobuf message:

```protobuf
syntax = "proto3";

package index_example;

message Foo {
  sint64 bar = 1;
  bool baz = 2;
}
```

Results in the following file:

```json
{
  "index": {
    "index_example.Foo": {
      "type": "message",
      "collection": "messages",
      "file": "test2.proto",
      "parent": ""
    },
    "index_example.Foo.bar": {
      "type": "field",
      "collection": "fields",
      "file": "test2.proto",
      "parent": "index_example.Foo"
    },
    "index_example.Foo.baz": {
      "type": "field",
      "collection": "fields",
      "file": "test2.proto",
      "parent": "index_example.Foo"
    }
  },
  "messages": {
    "index_example.Foo": {
      "name": "Foo",
      "full_name": "index_example.Foo",
      "description": "",
      "fields": [
        "index_example.Foo.bar",
        "index_example.Foo.baz"
      ],
      "messages": [],
      "enums": []
    }
  },
  "fields": {
    "index_example.Foo.bar": {
      "name": "bar",
      "full_name": "index_example.Foo.bar",
      "label": "LABEL_OPTIONAL",
      "type": "sint64",
      "full_type": "sint64",
      "description": ""
    },
    "index_example.Foo.baz": {
      "name": "baz",
      "full_name": "index_example.Foo.baz",
      "label": "LABEL_OPTIONAL",
      "type": "bool",
      "full_type": "bool",
      "description": ""
    }
  }
}
```


## The Collections

Besides `index`, the following arrays are exported at the root of the JSON document:

- `files`
- `services`
- `methods`
- `messages`
- `fields`
- `enums`
- `enum_values`

## Example

Given the following input file, `test.proto`:

```protobuf
syntax = "proto3";

import "google/protobuf/descriptor.proto";

package trinsic.protoc.gen.json.test;

// Test of a field extension
extend google.protobuf.FieldOptions {
  optional bool bool_option = 50000;
}

// Just a simple, hardworking enum
enum TestEnum {
  // Foo's value is 0. Foo indeed.
  FOO = 0;
  // Bar's value is 1. We don't like bar.
  BAR = 1;
}

// A message which is referenced as the field type for
// a field in another message
message TestReferencedMessage {
  // A string field
  string test_string_field = 1;
}

// A message
message TestMessage {
  // A message which is defined in another message
  message TestSubMessage {
    // A field in a message which is defined in another message
    // (...at the bottom of the sea)
    int64 test_sub_field = 1;
  }

  // A field with a type pointing to a message defined in another message.
  // This field also has a custom field option set on it.
  TestSubMessage test_sub_message_field = 1 [(bool_option) = false];

  // A field with a type pointing to a message defined externally
  TestReferencedMessage test_ref_field = 2;

  // A field with a primitive type
  int64 test_primitive_field = 3;

  // A field with an enum type
  TestEnum test_enum_field = 4;
}


// A message which is the input to a Method
message TestInputMessage {
  sint32 test_input_field = 1;
}

// A message which is the output of a Method
message TestOutputMessage {
  bool test_output_field = 2;
}

// A service
service TestService {
  // A method defined in a service
  rpc TestMethod              (TestInputMessage)             returns (TestOutputMessage);
}
```

Will generate the following file:

```json
{
  "index": {
    "trinsic.protoc.gen.json.test.TestEnum": {
      "type": "enum",
      "collection": "enums",
      "file": "test.proto",
      "parent": ""
    },
    "trinsic.protoc.gen.json.test.TestEnum.BAR": {
      "type": "enum_value",
      "collection": "enum_values",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestEnum"
    },
    "trinsic.protoc.gen.json.test.TestEnum.FOO": {
      "type": "enum_value",
      "collection": "enum_values",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestEnum"
    },
    "trinsic.protoc.gen.json.test.TestInputMessage": {
      "type": "message",
      "collection": "messages",
      "file": "test.proto",
      "parent": ""
    },
    "trinsic.protoc.gen.json.test.TestInputMessage.test_input_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestInputMessage"
    },
    "trinsic.protoc.gen.json.test.TestMessage": {
      "type": "message",
      "collection": "messages",
      "file": "test.proto",
      "parent": ""
    },
    "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage": {
      "type": "message",
      "collection": "messages",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestMessage"
    },
    "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage.test_sub_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_enum_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestMessage"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_primitive_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestMessage"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_ref_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestMessage"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_sub_message_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestMessage"
    },
    "trinsic.protoc.gen.json.test.TestOutputMessage": {
      "type": "message",
      "collection": "messages",
      "file": "test.proto",
      "parent": ""
    },
    "trinsic.protoc.gen.json.test.TestOutputMessage.test_output_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestOutputMessage"
    },
    "trinsic.protoc.gen.json.test.TestReferencedMessage": {
      "type": "message",
      "collection": "messages",
      "file": "test.proto",
      "parent": ""
    },
    "trinsic.protoc.gen.json.test.TestReferencedMessage.test_string_field": {
      "type": "field",
      "collection": "fields",
      "file": "test.proto",
      "parent": "trinsic.protoc.gen.json.test.TestReferencedMessage"
    },
    "trinsic.protoc.gen.json.test.TestService": {
      "type": "serviceProto",
      "collection": "services",
      "file": "test.proto",
      "parent": ""
    },
    "trinsic.protoc.gen.json.test.TestService.TestMethod": {
      "type": "methodProto",
      "collection": "methods",
      "file": "test.proto",
      "parent": ""
    }
  },
  "files": {
    "test.proto": {
      "name": "test.proto",
      "package": "trinsic.protoc.gen.json.test",
      "description": "",
      "services": [
        "trinsic.protoc.gen.json.test.TestService"
      ],
      "methods": [
        "trinsic.protoc.gen.json.test.TestService.TestMethod"
      ],
      "messages": [
        "trinsic.protoc.gen.json.test.TestReferencedMessage",
        "trinsic.protoc.gen.json.test.TestMessage",
        "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage",
        "trinsic.protoc.gen.json.test.TestInputMessage",
        "trinsic.protoc.gen.json.test.TestOutputMessage"
      ],
      "fields": [
        "trinsic.protoc.gen.json.test.TestReferencedMessage.test_string_field",
        "trinsic.protoc.gen.json.test.TestMessage.test_sub_message_field",
        "trinsic.protoc.gen.json.test.TestMessage.test_ref_field",
        "trinsic.protoc.gen.json.test.TestMessage.test_primitive_field",
        "trinsic.protoc.gen.json.test.TestMessage.test_enum_field",
        "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage.test_sub_field",
        "trinsic.protoc.gen.json.test.TestInputMessage.test_input_field",
        "trinsic.protoc.gen.json.test.TestOutputMessage.test_output_field"
      ],
      "enums": [
        "trinsic.protoc.gen.json.test.TestEnum"
      ],
      "enum_values": [
        "trinsic.protoc.gen.json.test.TestEnum.FOO",
        "trinsic.protoc.gen.json.test.TestEnum.BAR"
      ]
    }
  },
  "services": {
    "trinsic.protoc.gen.json.test.TestService": {
      "name": "TestService",
      "full_name": "trinsic.protoc.gen.json.test.TestService",
      "description": "A service",
      "methods": [
        "trinsic.protoc.gen.json.test.TestService.TestMethod"
      ]
    }
  },
  "methods": {
    "trinsic.protoc.gen.json.test.TestService.TestMethod": {
      "name": "TestMethod",
      "full_name": "trinsic.protoc.gen.json.test.TestService.TestMethod",
      "input_type": "trinsic.protoc.gen.json.test.TestInputMessage",
      "output_type": "trinsic.protoc.gen.json.test.TestOutputMessage",
      "description": "A method defined in a service"
    }
  },
  "messages": {
    "trinsic.protoc.gen.json.test.TestInputMessage": {
      "name": "TestInputMessage",
      "full_name": "trinsic.protoc.gen.json.test.TestInputMessage",
      "description": "A message which is the input to a Method",
      "fields": [
        "trinsic.protoc.gen.json.test.TestInputMessage.test_input_field"
      ],
      "messages": [],
      "enums": []
    },
    "trinsic.protoc.gen.json.test.TestMessage": {
      "name": "TestMessage",
      "full_name": "trinsic.protoc.gen.json.test.TestMessage",
      "description": "A message",
      "fields": [
        "trinsic.protoc.gen.json.test.TestMessage.test_sub_message_field",
        "trinsic.protoc.gen.json.test.TestMessage.test_ref_field",
        "trinsic.protoc.gen.json.test.TestMessage.test_primitive_field",
        "trinsic.protoc.gen.json.test.TestMessage.test_enum_field"
      ],
      "messages": [
        "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage"
      ],
      "enums": []
    },
    "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage": {
      "name": "TestSubMessage",
      "full_name": "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage",
      "description": "A message which is defined in another message",
      "fields": [
        "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage.test_sub_field"
      ],
      "messages": [],
      "enums": []
    },
    "trinsic.protoc.gen.json.test.TestOutputMessage": {
      "name": "TestOutputMessage",
      "full_name": "trinsic.protoc.gen.json.test.TestOutputMessage",
      "description": "A message which is the output of a Method",
      "fields": [
        "trinsic.protoc.gen.json.test.TestOutputMessage.test_output_field"
      ],
      "messages": [],
      "enums": []
    },
    "trinsic.protoc.gen.json.test.TestReferencedMessage": {
      "name": "TestReferencedMessage",
      "full_name": "trinsic.protoc.gen.json.test.TestReferencedMessage",
      "description": "A message which is referenced as the field type for\na field in another message",
      "fields": [
        "trinsic.protoc.gen.json.test.TestReferencedMessage.test_string_field"
      ],
      "messages": [],
      "enums": []
    }
  },
  "fields": {
    "trinsic.protoc.gen.json.test.TestInputMessage.test_input_field": {
      "name": "test_input_field",
      "full_name": "trinsic.protoc.gen.json.test.TestInputMessage.test_input_field",
      "label": "LABEL_OPTIONAL",
      "type": "sint32",
      "full_type": "sint32",
      "description": ""
    },
    "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage.test_sub_field": {
      "name": "test_sub_field",
      "full_name": "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage.test_sub_field",
      "label": "LABEL_OPTIONAL",
      "type": "int64",
      "full_type": "int64",
      "description": "A field in a message which is defined in another message\n(...at the bottom of the sea)"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_enum_field": {
      "name": "test_enum_field",
      "full_name": "trinsic.protoc.gen.json.test.TestMessage.test_enum_field",
      "label": "LABEL_OPTIONAL",
      "type": "TestEnum",
      "full_type": "trinsic.protoc.gen.json.test.TestEnum",
      "description": "A field with an enum type"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_primitive_field": {
      "name": "test_primitive_field",
      "full_name": "trinsic.protoc.gen.json.test.TestMessage.test_primitive_field",
      "label": "LABEL_OPTIONAL",
      "type": "int64",
      "full_type": "int64",
      "description": "A field with a primitive type"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_ref_field": {
      "name": "test_ref_field",
      "full_name": "trinsic.protoc.gen.json.test.TestMessage.test_ref_field",
      "label": "LABEL_OPTIONAL",
      "type": "TestReferencedMessage",
      "full_type": "trinsic.protoc.gen.json.test.TestReferencedMessage",
      "description": "A field with a type pointing to a message defined externally"
    },
    "trinsic.protoc.gen.json.test.TestMessage.test_sub_message_field": {
      "name": "test_sub_message_field",
      "full_name": "trinsic.protoc.gen.json.test.TestMessage.test_sub_message_field",
      "label": "LABEL_OPTIONAL",
      "type": "TestSubMessage",
      "full_type": "trinsic.protoc.gen.json.test.TestMessage.TestSubMessage",
      "description": "A field with a type pointing to a message defined in another message.\nThis field also has a custom field option set on it.",
      "options": {
        "trinsic.protoc.gen.json.test.bool_option": false
      }
    },
    "trinsic.protoc.gen.json.test.TestOutputMessage.test_output_field": {
      "name": "test_output_field",
      "full_name": "trinsic.protoc.gen.json.test.TestOutputMessage.test_output_field",
      "label": "LABEL_OPTIONAL",
      "type": "bool",
      "full_type": "bool",
      "description": ""
    },
    "trinsic.protoc.gen.json.test.TestReferencedMessage.test_string_field": {
      "name": "test_string_field",
      "full_name": "trinsic.protoc.gen.json.test.TestReferencedMessage.test_string_field",
      "label": "LABEL_OPTIONAL",
      "type": "string",
      "full_type": "string",
      "description": "A string field"
    }
  },
  "enums": {
    "trinsic.protoc.gen.json.test.TestEnum": {
      "name": "TestEnum",
      "full_name": "trinsic.protoc.gen.json.test.TestEnum",
      "description": "Just a simple, hardworking enum",
      "values": [
        "trinsic.protoc.gen.json.test.TestEnum.FOO",
        "trinsic.protoc.gen.json.test.TestEnum.BAR"
      ]
    }
  },
  "enum_values": {
    "trinsic.protoc.gen.json.test.TestEnum.BAR": {
      "name": "BAR",
      "full_name": "trinsic.protoc.gen.json.test.TestEnum.BAR",
      "description": "Bar's value is 1. We don't like bar.",
      "value": 1
    },
    "trinsic.protoc.gen.json.test.TestEnum.FOO": {
      "name": "FOO",
      "full_name": "trinsic.protoc.gen.json.test.TestEnum.FOO",
      "description": "Foo's value is 0. Foo indeed.",
      "value": 0
    }
  }
}
```