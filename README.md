# Flow Playground API

## Generating code from GQL

This project uses [gqlgen](https://github.com/99designs/gqlgen) to generate GraphQL server code from a GQL schema file.

```shell script
make generate
```

## Testing

```shell script
make test
```

## Running the server

```shell script
make run
```

When running locally, the GraphQL playground is available at [http://localhost:8080/](http://localhost:8080/).