// Example using the client in a typescript project

import { Client, loadClientWorker, MsgType, SetWorldType } from "../ts"; // "@momentum-xyz/posbus-client";

async function main() {
  const client = await loadClientWorker();
  const backendUrl = "http://localhost:8080";
  const token = "foo.bar.baz";
  const userId = "foobar";
  const worldId = "bazqux";
  const port = await client.connect(`${backendUrl}/posbus`, token, userId);
  port.onmessage = (ev) => {
    const [msgType, data] = ev.data;
    switch (msgType) {
      case MsgType.SET_WORLD: {
        const worldData = data; // Typescript type assertion works here.
        console.log("Entered world", worldData.name);
        break;
      }

      default:
    }
  };
  await client.teleport(worldId);
  await new Promise(f => setTimeout(f, 2345));
  port.postMessage([MsgType.USER_ACTION, { foo: "bar" }]);
}
