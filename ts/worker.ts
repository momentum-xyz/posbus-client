import "../build/wasm_exec";
import wasmUrl from "../build/pbc.wasm";
import { PostMessageType } from "./worker_messaging";

// Exported from above wasm
declare const PBC: {
  connect: (url: string, token: string, userId: string) => Promise<void>;
  disconnect: () => void;
  setPort: (port: MessagePort) => void;
  teleport: (world: string) => void;
  send: (msgType: string, data: any) => void;
};

let msgPort: MessagePort | null = null;

onmessage = async (e: MessageEvent) => {
  switch (e.data.type) {
    case PostMessageType.WORKER_LOAD: {
      const { go, instance } = await loadWasm();
      void go.run(instance);
      e.ports[0]?.postMessage(true);
      break;
    }
    case PostMessageType.MSG_PORT: {
      msgPort = e.ports?.[0] ?? null;
      if (msgPort != null) {
        PBC.setPort(msgPort);
        msgPort.onmessage = (ev) => {
          const [msgType, data] = ev.data;
          // TODO: avoid the stringify
          PBC.send(msgType, JSON.stringify(data));
        };
      }
      break;
    }
    case PostMessageType.CONNECT: {
      const { url, token, userId } = e.data;
      try {
        await PBC.connect(url, token, userId);
        e.ports[0]?.postMessage(true);
      } catch (err) {
        e.ports[0]?.postMessage({ type: PostMessageType.ERROR, err });
      }
      break;
    }
    case PostMessageType.DISCONNECT: {
      PBC.disconnect();
      break;
    }
    case PostMessageType.TELEPORT: {
      const { world } = e.data;
      PBC.teleport(world);
      break;
    }
    default:
      console.warn("Unknown message", e);
  }
};

interface LoadedWasm {
  go: Go;
  instance: WebAssembly.Instance;
}

export const loadWasm = async (
  url = new URL(`./${wasmUrl}`, import.meta.url)
): Promise<LoadedWasm> => {
  const go = new Go();
  const { instance } = await WebAssembly.instantiateStreaming(
    fetch(url),
    go.importObject
  );
  return { go, instance };
};
