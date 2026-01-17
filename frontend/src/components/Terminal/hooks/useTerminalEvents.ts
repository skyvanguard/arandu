import type React from "react";
import { useEffect } from "react";
import { Terminal as XTerminal } from "@xterm/xterm";

// Type for XTerm event handlers
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type XTermEventHandler = ((...args: any[]) => void) | undefined;

type XTermEventName =
  | "onBell"
  | "onBinary"
  | "onCursorMove"
  | "onData"
  | "onKey"
  | "onLineFeed"
  | "onRender"
  | "onResize"
  | "onScroll"
  | "onSelectionChange"
  | "onTitleChange"
  | "onWriteParsed";

/**
 * Hook interno para vincular un handler a un evento del terminal
 */
export function useBind(
  termRef: React.RefObject<XTerminal | null>,
  handler: XTermEventHandler,
  eventName: XTermEventName,
): void {
  useEffect(() => {
    if (!termRef.current || typeof handler !== "function") return;

    const term = termRef.current;
    const eventBinding = term[eventName](handler);

    return () => {
      if (!eventBinding) return;
      eventBinding.dispose();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [handler]);
}

export interface TerminalEventHandlers {
  onBell?: () => void;
  onBinary?: (data: string) => void;
  onCursorMove?: () => void;
  onData?: (data: string) => void;
  onKey?: (key: { key: string; domEvent: KeyboardEvent }) => void;
  onLineFeed?: () => void;
  onRender?: () => void;
  onResize?: (cols: number, rows: number) => void;
  onScroll?: (ydisp: number) => void;
  onSelectionChange?: () => void;
  onTitleChange?: (title: string) => void;
  onWriteParsed?: (data: string) => void;
}

/**
 * Hook para vincular todos los event handlers del terminal
 */
export function useTerminalEvents(
  xtermRef: React.RefObject<XTerminal | null>,
  handlers: TerminalEventHandlers,
): void {
  useBind(xtermRef, handlers.onBell, "onBell");
  useBind(xtermRef, handlers.onBinary, "onBinary");
  useBind(xtermRef, handlers.onCursorMove, "onCursorMove");
  useBind(xtermRef, handlers.onData, "onData");
  useBind(xtermRef, handlers.onKey, "onKey");
  useBind(xtermRef, handlers.onLineFeed, "onLineFeed");
  useBind(xtermRef, handlers.onRender, "onRender");
  useBind(xtermRef, handlers.onResize, "onResize");
  useBind(xtermRef, handlers.onScroll, "onScroll");
  useBind(xtermRef, handlers.onSelectionChange, "onSelectionChange");
  useBind(xtermRef, handlers.onTitleChange, "onTitleChange");
  useBind(xtermRef, handlers.onWriteParsed, "onWriteParsed");
}
