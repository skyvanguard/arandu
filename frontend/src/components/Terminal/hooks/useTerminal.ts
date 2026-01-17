import type React from "react";
import { useEffect, useRef } from "react";
import { ITerminalAddon, Terminal as XTerminal } from "@xterm/xterm";
import { CanvasAddon } from "@xterm/addon-canvas";
import { FitAddon } from "@xterm/addon-fit";
import { Unicode11Addon } from "@xterm/addon-unicode11";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { WebglAddon } from "@xterm/addon-webgl";

// Terminal theme configuration
export const terminalTheme = {
  background: "#1e1e1e",
  foreground: "#cccccc",
  cursor: "#ffffff",
  cursorAccent: "#1e1e1e",
  selectionBackground: "#264f78",
  black: "#000000",
  red: "#cd3131",
  green: "#0dbc79",
  yellow: "#e5e510",
  blue: "#2472c8",
  magenta: "#bc3fbc",
  cyan: "#11a8cd",
  white: "#e5e5e5",
  brightBlack: "#666666",
  brightRed: "#f14c4c",
  brightGreen: "#23d18b",
  brightYellow: "#f5f543",
  brightBlue: "#3b8eea",
  brightMagenta: "#d670d6",
  brightCyan: "#29b8db",
  brightWhite: "#ffffff",
};

const isWebGl2Supported = !!document
  .createElement("canvas")
  .getContext("webgl2");

// Create addons for terminal
function createAddons(): ITerminalAddon[] {
  return [
    new Unicode11Addon(),
    new CanvasAddon(),
    isWebGl2Supported ? new WebglAddon() : new WebLinksAddon(),
  ];
}

export interface UseTerminalOptions {
  id?: string;
  customKeyEventHandler?: (event: KeyboardEvent) => boolean;
  onInit?: (term: XTerminal) => void;
}

export interface UseTerminalResult {
  divRef: React.RefObject<HTMLDivElement>;
  xtermRef: React.RefObject<XTerminal | null>;
}

/**
 * Hook para inicializar y manejar una instancia de XTerm
 */
export function useTerminal(options: UseTerminalOptions): UseTerminalResult {
  const { id, customKeyEventHandler, onInit } = options;
  const divRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<XTerminal | null>(null);

  // Initialize terminal
  useEffect(() => {
    if (!divRef.current || xtermRef.current) return;

    const xterm = new XTerminal({
      convertEol: true,
      allowProposedApi: true,
      theme: terminalTheme,
    });

    // Load addons
    const addons = createAddons();
    addons.forEach((addon) => {
      xterm.loadAddon(addon);
    });

    const fitAddon = new FitAddon();
    xterm.loadAddon(fitAddon);

    // Add custom key event handler if provided
    if (customKeyEventHandler) {
      xterm.attachCustomKeyEventHandler(customKeyEventHandler);
    }

    xtermRef.current = xterm;
    xterm.open(divRef.current);
    fitAddon.fit();
  }, [id, customKeyEventHandler]);

  // Handle onInit callback
  useEffect(() => {
    if (!xtermRef.current) return;
    if (typeof onInit !== "function") return;
    onInit(xtermRef.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [xtermRef.current, onInit]);

  return { divRef, xtermRef };
}
