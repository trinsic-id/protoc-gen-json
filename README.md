# protoc-gen-json

A `protoc` plugin which generates a JSON representation of protobuf files.

This is used by Trinsic to enable generation of rich documentation for our API objects, which are defined in proto files.

## Usage

As with all `protoc` plugins, its usage is simple. 

#### Placing in your PATH

The simplest option: simply place `protoc-gen-json[.exe]` somewhere that your PATH can see.

Then, invoke it:

```bash
protoc --json_out={OUTPUT_DIRECTORY} --json_opt={OUTPUT_FILENAME} [standard protoc flags follow]
```

#### Specifying path to plugin during protoc instantiation

```bash
protoc --plugin=protoc-gen-json={PATH_TO_EXECUTABLE} --json_out={OUTPUT_DIRECTORY} --json_opt={OUTPUT_FILENAME} [standard protoc flags follow]
```


#### Plugin Config

`--json_out` should point to the output _directory_, where you would like your output JSON file to be saved.

`--json_opt` (_optional_) is the output filename. Defaults to `proto.json` if not specified.

For example, `--json_out=/foo/bar --json_opt=baz.json` will write the output file to `/foo/bar/baz.json`.


## Output Format

See [OUTPUT.md](/OUTPUT.md) for documentation about the output format.

## Known Issues / Todo

- [ ] Full Custom Option Support
    - [ ] Handle Options on all non-Field objects
    - [ ] Handle non-primitive Option types
    - [ ] Handle `bytes` Option types
- [ ] Improved label handling (repeated/optional/etc.)
- [ ] Add default value support