import { PostMessageType, workerCall } from "./worker_messaging";

const WORKER_FILE = "worker.mjs"; // TODO: can we import this (like .wasm)?
/**
 * Momentum posbus client.
 *
 * Javascript wrapper around the WASM client.
 */
export class Client {
  constructor(private readonly worker: Worker) {}

  async connect(url: string, token: string, userId: string): Promise<MessagePort> {
    const { port1, port2 } = new MessageChannel();
    this.worker.postMessage({ type: PostMessageType.MSG_PORT }, [port2]);
    await workerCall(this.worker, {
      type: PostMessageType.CONNECT,
      url,
      token,
      userId,
    });
    return port1;
  }

  async disconnect(): Promise<void> {
    this.worker.postMessage({ type: PostMessageType.DISCONNECT });
  }

  async teleport(worldId: string): Promise<void> {
    this.worker.postMessage({ type: PostMessageType.TELEPORT, world: worldId });
  }
}

export const loadClientWorker = async (
  url = new URL(`./${WORKER_FILE}`, import.meta.url),
  wasmUrl?: URL
): Promise<Client> => {
  const worker = new Worker(url, { type: "module", name: "PBC" });
  await workerCall(worker, { type: PostMessageType.WORKER_LOAD, wasmUrl: wasmUrl?.toString() });
  return new Client(worker);
};
