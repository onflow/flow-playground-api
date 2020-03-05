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

### Running with Datastore emulator

Install the [Google Cloud Datastore Emulator](https://cloud.google.com/datastore/docs/tools/datastore-emulator). 

Start the Datastore emulator:
```shell script
gcloud beta emulators datastore start
```

In a separate process, run the server with the `run-datastore` target:

```shell script
make run-datastore
```

### Configuration options

The following environment variables can be used to configure the API. Default values are shown below:

```shell script
FLOW_PORT=8080
FLOW_DEBUG=false
FLOW_ALLOWEDORIGINS="http://localhost:3000"

FLOW_SESSIONAUTHKEY="428ce08c21b93e5f0eca24fbeb0c7673"
FLOW_SESSIONMAXAGE="157680000s"
FLOW_SESSIONCOOKIESSECURE=true
FLOW_SESSIONCOOKIESHTTPONLY=true
FLOW_SESSIONCOOKIESSAMESITENONE=false

FLOW_LEDGERCACHESIZE=128
FLOW_STORAGEBACKEND="memory"

FLOW_DATASTORE_GCPPROJECTID
FLOW_DATASTORE_TIMEOUT="5s"
```