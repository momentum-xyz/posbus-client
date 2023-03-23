import type * as msg from "../build/channel_types";

/**
 * The actual Postbus messages are send through postMessage/onmessage, encapsulated inside a tuple to pass along its type.
 *
 * This uses the (from go) generated posbus.d.ts and wraps these for use by the Client.
 *
 * They are wrapped with array or else all the messages need to get a 'type' field, which they don't have on the golang side.
 * This also avoids conflict, if the message has a field names type itself (since we have generic messages).
 */

export interface PosbusEvent extends MessageEvent {
  data: msg.PosbusMessage;
}

export interface PosbusPort extends MessagePort {
  onmessage: ((this: MessagePort, ev: PosbusEvent) => any) | null;
  postMessage: (message: msg.PosbusMessage) => void;
}

export type * as posbus from "../build/posbus";
