const WS_BASE = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080";

type EventHandler<T = unknown> = (payload: T) => void;

interface InMessage {
  type: string;
  payload: unknown;
}

let socket: WebSocket | null = null;
const listeners = new Map<string, Set<EventHandler>>();

function emit(eventName: string, payload: unknown) {
  const handlers = listeners.get(eventName);
  if (handlers) handlers.forEach((h) => h(payload));
}

export function connectWS(token: string) {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return;

  socket = new WebSocket(`${WS_BASE}/ws?token=${token}`);

  socket.onopen = () => emit("connect", null);
  socket.onclose = () => emit("disconnect", null);
  socket.onerror = () => emit("ws-error", null);

  socket.onmessage = (event) => {
    try {
      const msg: InMessage = JSON.parse(event.data as string);
      emit(msg.type, msg.payload);
    } catch {
      /* ignore malformed messages */
    }
  };
}

export function onWS<T = unknown>(eventName: string, handler: EventHandler<T>) {
  if (!listeners.has(eventName)) listeners.set(eventName, new Set());
  listeners.get(eventName)!.add(handler as EventHandler);
}

export function offWS(eventName: string, handler?: EventHandler) {
  if (!handler) {
    listeners.delete(eventName);
    return;
  }
  listeners.get(eventName)?.delete(handler);
}

export function sendWS(type: string, payload: Record<string, unknown>) {
  if (socket?.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type, ...payload }));
  }
}

export function isConnectedWS(): boolean {
  return socket?.readyState === WebSocket.OPEN;
}

export function disconnectWS() {
  socket?.close();
  socket = null;
  listeners.clear();
}
