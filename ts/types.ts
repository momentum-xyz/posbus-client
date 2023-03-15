import type * as posbus from "../build/posbus";

/**
 * The actual Postbus messages are send through postMessage/onmessage, encapsulated inside a tuple to pass along its type.
 *
 * This uses the (from go) generated posbus.d.ts and wraps these for use by the Client.
 *
 * They are wrapped with array or else all the messages need to get a 'type' field, which they don't have on the golang side.
 * This also avoids conflict, if the message has a field names type itself (since we have generic messages).
 */

// TODO: auto generate this:
export enum MsgType {
  // Incoming and outgoing types:
  // None yet...

  // Incoming types:
  SIGNAL = "signal",
  SET_WORLD = "set_world",
  ADD_OBJECTS = "add_objects",
  SET_USER_TRANSFORM = "set_users_transforms",

  // Outgoing types:
  SEND_TRANSFORM = "send_transform",
  USER_ACTION = "user_action",
}

// Incoming!
export type SignalType = [MsgType.SIGNAL, posbus.Signal]
export type SetWorldType = [MsgType.SET_WORLD, posbus.SetWorldData]
export type AddObjectsType = [MsgType.ADD_OBJECTS, posbus.AddObjects]
export type SetUsersTransforms = [MsgType.SET_USER_TRANSFORM, posbus.SetUsersTransforms]

export type IncomingMessage = SignalType | SetWorldType | AddObjectsType | SetUsersTransforms;

// Outgoing!
export type PositionTransformType = [MsgType.SEND_TRANSFORM, posbus.UserTransform]
export type UserActionType = [MsgType.USER_ACTION, Record<string, unknown>];  // TODO

export type PostbusMessage = PositionTransformType | UserActionType;


export interface PosbusEvent extends MessageEvent {
  data: IncomingMessage;
}

export interface PosbusPort extends MessagePort {
  onmessage: ((this: MessagePort, ev: PosbusEvent) => any) | null;
  postMessage: (message: PostbusMessage) => void;
}
