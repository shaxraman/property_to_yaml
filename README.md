# property_to_yaml

A utility for converting `.properties` files (Java-style key=value) into structured YAML.

## Features

- Supports nested structures (dot-separated keys become nested YAML objects)
- Preserves key order as in the original file, or sorts alphabetically (optional)
- Optionally inserts blank lines between top-level YAML blocks for readability ("pretty" mode)
- Simple usage — just Go, no external libraries for properties parsing

## Usage

```sh
go run main.go [options] <file.properties>
```

### Examples

- Convert while preserving the original order (default):

  ```sh
  go run main.go stage.mnru-loans.properties
  ```

- Convert with alphabetical sorting on all levels:

  ```sh
  go run main.go --sort alpha stage.mnru-loans.properties
  ```

- Pretty YAML (blank line between top-level blocks):

  ```sh
  go run main.go --pretty stage.mnru-loans.properties
  ```

- All options together:

  ```sh
  go run main.go --pretty --sort alpha stage.mnru-loans.properties
  ```

### Options

- `--sort original` — preserve key order as in the `.properties` file (default)
- `--sort alpha` — sort all keys alphabetically
- `--pretty` — insert a blank line before each top-level YAML block

## Example

Source `example.properties` file:

```properties
db.host=localhost
db.port=5432

rabbit.enabled=true
rabbit.host=localhost
```

Resulting YAML (with --pretty):

```yaml
db:
  host: localhost
  port: 5432

rabbit:
  enabled: true
  host: localhost
```

## Requirements

- Go 1.17+ (latest version recommended)

## Install dependencies

```sh
go mod tidy
```

## TODO / Roadmap

- [ ] Preserve and transfer comments from properties to YAML
- [ ] Support for arrays and multiline values (if needed)
- [ ] More flexible output customization flags
