import "websocket-polyfill";
// import { PBClient } from "@momentum-xyz/posbus-client";
import { PBClient, MsgType } from "../../dist/index.mjs";
import fs from "fs";

const [userId, token] = process.argv.slice(2);

if (!userId || !token) {
  console.error("Usage: node index.js <userId> <token>");
  process.exit(1);
}

const buffer = fs.readFileSync(
  // "./node_modules/@momentum-xyz/posbus-client/dist/pbc.wasm"
  "../../dist/pbc.wasm"
);

async function main(userId, token) {
  const client = new PBClient((event) => {
    console.log(`PosBus message [${userId}]:`, event.data);
  });

  await client.load(buffer);

  console.log(`PosBus client loaded [${userId}]`, client);
  await client.connect(`https://dev.odyssey.ninja/posbus`, token, userId);
  // await client.connect(`http://localhost:4000/posbus`, token, userId);
  // await client.connect(`ws://localhost:4000/posbus`, token, userId);

  await sleep(500);
  console.log(`PosBus client teleport [${userId}]`);
  await client.teleport("00000000-0000-8000-8000-000000000005");

  await sleep(3000);

  console.log(`PosBus client send MY_TRANSFORM [${userId}]`);
  await client.send(MsgType.MY_TRANSFORM, {
    position: { x: 0, y: 0, z: 5 },
    rotation: { x: 0, y: 0, z: 0 },
  });
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

main(userId, token).catch((err) => {
  console.error(err);
});
