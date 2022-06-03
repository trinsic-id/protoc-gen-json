# protoc-gen-json

A `protoc` plugin which generates a JSON representation of protobuf files.

This is used by Trinsic to enable generation of rich documentation for our API objects, which are defined in proto files.

## Usage

As with all `protoc` plugins, its usage is simple. You have two options.

#### Placing in your PATH

The simplest option: simply place `protoc-gen-json[.exe]` somewhere that your PATH can see.

Then, invoke it:

