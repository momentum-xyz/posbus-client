/**
 * A little mini protocol to communicate between worker and main window.
 */

/**
 * Types for postMessage communication.
 */
export const enum PostMessageType {
  WORKER_LOAD = "PBC_LOAD", // Indicate worker should start loading/initialising.
  ERROR = "PBC_ERR", // Indicate some error send back to main.
  CONNECT = "PBC_CONN", // Indicate connection should be made.
  DISCONNECT = "PBC_DISC", // Indicate connection should be closed.
  MSG_PORT = "PBC_PORT", // Message to send communication port to worker.
  TELEPORT = "PBC_TP", // Teleport to a world.
}

/**
 * Async message to a worker with a response.
 *
 * Wrap postMessage to get an async request-response mechanism.
 */
export async function workerCall(
  worker: Worker | ServiceWorker,
  data: any
): Promise<any> {
  return new Promise((resolve, reject) => {
    const { port1, port2 } = new MessageChannel();
    port1.onmessage = (event: MessageEvent) => {
      if (event.data.type === PostMessageType.ERROR) {
        reject(event.data.err);
      } else {
        resolve(event.data);
      }
    };
    worker.postMessage(data, [port2]);
  });
}
