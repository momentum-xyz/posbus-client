import "websocket-polyfill";
import { PBClient } from "@momentum-xyz/posbus-client";
// import { PBClient } from "../../dist/index.mjs";
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

async function main(userId: string, token: string) {
  try {
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
  } catch (err) {
    console.error(err);
  }
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

main(userId, token);
