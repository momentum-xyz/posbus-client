import type * as msg from "../build/channel_types";

/**
 * The actual Postbus messages are send through postMessage/onmessage, encapsulated inside a tuple to pass along its type.
 *
 * This uses the (from go) generated posbus.d.ts and wraps these for use by the Client.
 *
 * They are wrapped with array or else all the messages need to get a 'type' field, which they don't have on the golang side.
 * This also avoids conflict, if the message has a field names type itself (since we have generic messages).
 */

// Incoming!

export type IncomingMessage =
  | msg.SignalType
  | msg.SetWorldType
  | msg.MyTransformType
  | msg.AddUsersType
  | msg.AddObjectsType
  | msg.UsersTransformListType;

// Outgoing!
// export type PositionTransformType = [MsgType.MY_TRANSFORM, posbus.MyTransform]
// export type UserActionType = [MsgType.USER_ACTION, Record<string, unknown>];  // TODO

export type PostbusMessage = msg.ObjectPositionType | msg.GenericMessageType;

export interface PosbusEvent extends MessageEvent {
  data: IncomingMessage;
}

export interface PosbusPort extends MessagePort {
  onmessage: ((this: MessagePort, ev: PosbusEvent) => any) | null;
  postMessage: (message: PostbusMessage) => void;
}


export type * as posbus from "../build/posbus";
