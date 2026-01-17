// This terminal is a combination of the following packages:
// https://gist.github.com/mastersign/90d0ab06f040092e4ca27a3b59820cb9
// https://github.com/reubenmorgan/xterm-react/blob/6c8bb143387a6abc35ff54a3e099c46e5be8819c/src/Xterm.tsx
import React from "react";
import { ITerminalOptions, Terminal as XTerminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";

import dockerSvg from "@/assets/docker.svg";
import { Log } from "@/generated/graphql";

import { headerStyles } from "./Terminal.css";
import {
  useTerminal,
  useTerminalLogs,
  useTerminalEvents,
  TerminalEventHandlers,
} from "./hooks";

type XTermProps = TerminalEventHandlers & {
  customKeyEventHandler?(event: KeyboardEvent): boolean;
  className?: string;
  id?: string;
  onDispose?: (term: XTerminal) => void;
  onInit?: (term: XTerminal) => void;
  options?: ITerminalOptions;
  status?: string;
  title?: React.ReactNode;
  logs?: Log[];
  isRunning?: boolean;
};

export const Terminal = ({
  id,
  className,
  customKeyEventHandler,
  onInit,
  title,
  logs = [],
  isRunning = false,
  // Event handlers
  onBell,
  onBinary,
  onCursorMove,
  onData,
  onKey,
  onLineFeed,
  onRender,
  onResize,
  onScroll,
  onSelectionChange,
  onTitleChange,
  onWriteParsed,
}: XTermProps) => {
  // Initialize terminal
  const { divRef, xtermRef } = useTerminal({
    id,
    customKeyEventHandler,
    onInit,
  });

  // Handle logs
  useTerminalLogs({
    xtermRef,
    id,
    logs,
  });

  // Bind event handlers
  useTerminalEvents(xtermRef, {
    onBell,
    onBinary,
    onCursorMove,
    onData,
    onKey,
    onLineFeed,
    onRender,
    onResize,
    onScroll,
    onSelectionChange,
    onTitleChange,
    onWriteParsed,
  });

  return (
    <>
      <div className={headerStyles}>
        {isRunning ? (
          <>
            <img src={dockerSvg} alt="Docker" width="14" height="14" />
            {title} - Active
          </>
        ) : (
          "Disconnected"
        )}
      </div>
      <div id={id} className={className} ref={divRef} />
    </>
  );
};
