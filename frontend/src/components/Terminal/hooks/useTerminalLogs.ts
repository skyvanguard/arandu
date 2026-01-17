import type React from "react";
import { useEffect, useRef } from "react";
import { Terminal as XTerminal } from "@xterm/xterm";

export interface TerminalLog {
  id: string;
  text: string;
}

export interface UseTerminalLogsOptions {
  xtermRef: React.RefObject<XTerminal | null>;
  id?: string;
  logs?: TerminalLog[];
}

/**
 * Hook para manejar los logs del terminal
 * Escribe nuevos logs al terminal y limpia cuando cambia el ID
 */
export function useTerminalLogs(options: UseTerminalLogsOptions): void {
  const { xtermRef, id, logs = [] } = options;
  const renderedLogIds = useRef<string[]>([]);

  // Clear terminal when ID changes
  useEffect(() => {
    if (!xtermRef.current) return;

    xtermRef.current.clear();
    renderedLogIds.current = [];
  }, [id, xtermRef]);

  // Write new logs to terminal
  useEffect(() => {
    if (!xtermRef.current) return;

    for (const log of logs) {
      if (renderedLogIds.current.includes(log.id)) continue;

      xtermRef.current.writeln(log.text);
      renderedLogIds.current.push(log.id);
    }
  }, [logs, xtermRef]);
}
