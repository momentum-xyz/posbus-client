import "../build/wasm_exec";
import wasmUrl from "../build/pbc.wasm";
import type { PosbusEvent, PosbusPort } from "./types";

declare const PBC: {
  connect: (url: string, token: string, userId: string) => Promise<void>;
  disconnect: () => void;
  setPort: (port: MessagePort) => void;
  teleport: (world: string) => void;
  send: (msgType: string, data: any) => void;
};

interface LoadedWasm {
  go: Go;
  instance: WebAssembly.Instance;
}

export class PBClient {
  public onMessageCallback: ((event: PosbusEvent) => void) | undefined;

  constructor(onMessage?: (event: PosbusEvent) => void) {
    this.onMessageCallback = onMessage;
  }

  async load(buffer?: ArrayBuffer) {
    const { go, instance } = await this.loadWasm(buffer);
    go.run(instance)
      .then(() => {
        console.info("go stopped");
      })
      .catch((err) => {
        console.error("go run", err);
      });
    // need delay?
    if (typeof PBC === "undefined") {
      throw new Error("PBC undefined");
    }
    this.pbc = PBC;
  }

  async connect(
    url: string,
    token: string,
    userId: string
  ): Promise<PosbusPort> {
    const { port1, port2 } = new MessageChannel();
    if (this.onMessageCallback) {
      port1.onmessage = this.onMessageCallback;
    }
    this._getPBC().setPort(port2);
    await this._getPBC().connect(url, token, userId);
    return port1;
  }

  disconnect() {
    this._getPBC().disconnect();
  }

  teleport(world: string) {
    this._getPBC().teleport(world);
  }

  send(msgType: string, data: any) {
    this._getPBC().send(msgType, JSON.stringify(data));
  }

  private pbc: typeof PBC | null = null;
  private _getPBC(): typeof PBC {
    if (!this.pbc) throw new Error("PBC not loaded");
    return this.pbc;
  }

  private async loadWasm(buffer?: ArrayBuffer): Promise<LoadedWasm> {
    const go = new Go();

    if (buffer) {
      const { instance } = await WebAssembly.instantiate(
        buffer,
        go.importObject
      );
      return { go, instance };
    }

    const url = new URL(`./${wasmUrl}`, import.meta.url);
    const { instance } = await WebAssembly.instantiateStreaming(
      fetch(url),
      go.importObject
    );
    return { go, instance };
  }
}
