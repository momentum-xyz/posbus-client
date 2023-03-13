# Momentum posbus client

Client library to connect to Momentum websocket backend.

The Momentum backend exposes a websocket with a custom binary protocol. This library is for client to connect and talk the custom protocol.

Currently there is a client for Go and Javascript (in browser).

## Usage

**TODO**: this is only initial version, not fully functional and API not definitive yet.

```shell
npm install @momentum-xyz/posbus-client
```

```typescript
import { loadClientWorker } from "@momentum-xyz/posbus-client";

const client = await loadClientWorker();
// set backendUrl and authenticate here
let port = await client.connect(`${backendUrl}/posbus`), token, userId);
port.onmessage = (msg) => {
    // handle incoming messages
}
// select a world
client.teleport(worldId);
// talk back
port.postMessage(msgType, data);
```

## Development

This is a mixed Go and Typescript project.

### TL;DR

```shell
git clone git@github.com:momentum-xyz/posbus-client.git
cd posbus-client
npm install
npm start
```

The output of `npm start` should be an URL to see the example webpage.

### Prerequisites

- [Go](https://go.dev/), latest stable release.
- [Node](https://nodejs.org/), latest LTS release.
- [Make](https://www.gnu.org/software/make/)

### Building

Building is done with Go and [esbuild](https://esbuild.github.io/) for the typescript files.

As is common with Go projects, the build scripts use `make`, see [Makefile](./Makefile) for all the commands.

As is common with Node projects, the build scripts are in [package.json](./package.json).

To be able to integrate the javascript building with building Go parts the esbuild configuration in written in Go, see [build_js](./cmd/build_js/main.go).
Esbuild doesn't output typescript definition files, so `tsc` is used for that and tsconfig.json is setup to only output the declarations (since the rest is handled by esbuild).

### Running

There is an javascript example to run:

```shell
make clean run-example
```

### Project structure

The base is a Go project, with some added typescript bits.

* _cmd_: Source of Go binaries
* _docs_: Documentation
* _example_: Examples of usages
* _pbc_: Go package of the client library
* _ts_: Typescript source if the javascript library

Generated directories:

* _build_: Output of (intermediate) build files
* _dist_: Output of files to distribute
* _node_modules_: NPM depedencies


### How it works

See [Architecture.md](./docs/architecture.md)
