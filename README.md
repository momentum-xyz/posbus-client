# Momentum posbus client

Client library to connect to Momentum websocket backend.

The Momentum backend exposes a websocket with a custom binary protocol. This library is for client to connect and talk the custom protocol.

Currently there is a client for Go and Javascript (in browser).

## Usage

**TODO**: this is only initial version, not fully functional and API not definitive yet.

For now the packages are only hosted on Github npm package repository.
To use this you need to [authenticate](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-npm-registry#authenticating-to-github-packages).
Read the Github documentation, create PAT and:

```shell
npm login --scope=@momentum-xyz --auth-type=legacy --registry=https://npm.pkg.github.com
```

Install the package in your project:

```shell
npm install @momentum-xyz/posbus-client
```

The library can be used from browser or nodejs but in a different manner.

### Browser

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

### Nodejs

```typescript
import "websocket-polyfill"; // not supported natively in nodejs
import {
  posbus,
  PBClient,
  PosbusEvent,
  PosbusPort,
  MsgType,
} from "@momentum-xyz/posbus-client";
import fs from "fs";

const wasmURL = require.resolve("@momentum-xyz/posbus-client/pbc.wasm");
const wasmPBC = fs.readFileSync(wasmURL);

const { BACKEND_URL = "https://dev.odyssey.ninja" } = process.env;
const POSBUS_URL = `${BACKEND_URL}/posbus`;

async function main(userId: string, token: string) {
  // Instanciate the client
  const client = new PBClient((event) => {
    console.log(`PosBus message [${userId}]:`, event.data);
    // handle incoming messages
  });

  await client.load(wasmPBC);

  await client.connect(POSBUS_URL, token, userId);

  // select a world
  await client.teleport(worldId);

  // send a message
  await client.send(MsgType.MY_TRANSFORM, {
    position: { x: 0, y: 0, z: 5 },
    rotation: { x: 0, y: 0, z: 0 },
  });
}

// get token and userId from auth
main(userId, token).catch((err) => {
  console.error(err);
});
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

As is common with Go projects, the build scripts use `make`, run `make help` for all the commands or see [Makefile](./Makefile).

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

- _cmd_: Source of Go binaries
- _docs_: Documentation
- _example_: Examples of usages
- _pbc_: Go package of the client library
- _ts_: Typescript source if the javascript library
- _test_: Integration tests

Generated directories:

- _build_: Output of (intermediate) build files
- _dist_: Output of files to distribute
- _node_modules_: NPM dependencies
- _test-results_: Output from test runs

### Typescript/javascript

When developing the typescript parts, to make your editor happy, you need the generated files (in the build directory).
An `npm install` will call the prepare run script that calls the build to get you started. To pick up changes, run `make js` when needed.

### Go

The go client depends on the controller, to keep that up to date (to 'pull' in the latest changes) run `GOPROXY=direct go get -u github.com/momentum-xyz/ubercontroller/pkg/posbus@develop`

### Testing

Prerequisites: The integration tests use containers, so a configured [docker](https://www.docker.com) environment is required. Otherwise run go test with `-short` to skip these.

```
make test
```

### Releasing

Git tags following [semver](https://semver.org/) are used and handled on CI/CD.

Example: publish a prerelease version on Github packages:

```
npm version prerelease
git push --tags
```

Should work if your local git is setup to properly all sign commits and push tags by default.

### How it works

See [Architecture.md](./docs/architecture.md)
