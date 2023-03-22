// Example using the client in a typescript project

import { loadClientWorker, MsgType } from "@momentum-xyz/posbus-client";
import type { posbus } from "@momentum-xyz/posbus-client";

async function main(): Promise<void> {
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
      case MsgType.MY_TRANSFORM: {
          const {
            id: userId,
            transform: { location, rotation },
          } = data;
          console.log(
            `My user ${userId} ⊹${fmtVec3(location)} ∡${fmtVec3(rotation)}`
          );
        break;
      }
      case MsgType.USERS_TRANSFORM_LIST: {
        for (const uTransform of data.value) {
          const {
            id: userId,
            transform: { location, rotation },
          } = uTransform;
          console.log(
            `User ${userId} ⊹${fmtVec3(location)} ∡${fmtVec3(rotation)}`
          );
        }
        break;
      }
      default:
    }
  };
  await client.teleport(worldId);
  await new Promise((resolve) => setTimeout(resolve, 2345));
  port.postMessage([MsgType.GENERIC_MESSAGE, { Topic: "foo", Data: "bar" }]);
}

const fmtVec3 = ({ x, y, z }: posbus.Vec3): string => `${x},${y},${z}`;

export default main;
