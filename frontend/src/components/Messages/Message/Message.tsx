import type React from "react";
import clsx from "clsx";
import { formatDistanceToNowStrict } from "date-fns";
import { useState } from "react";

import logoPng from "@/assets/logo.png";
import mePng from "@/assets/me.png";
import { Button } from "@/components/Button/Button";
import { Icon } from "@/components/Icon/Icon";
import { TaskStatus, TaskType } from "@/generated/graphql";

import {
  avatarStyles,
  contentStyles,
  iconStyles,
  messageStyles,
  outputStyles,
  rightColumnStyles,
  timeStyles,
  wrapperStyles,
} from "./Message.css";

const taskTypeIcons: Record<TaskType, React.ReactNode> = {
  [TaskType.Browser]: <Icon.Browser />,
  [TaskType.Terminal]: <Icon.Terminal />,
  [TaskType.Code]: <Icon.Code />,
  [TaskType.Ask]: <Icon.MessageQuestion />,
  [TaskType.Done]: <Icon.CheckCircle />,
  [TaskType.Input]: null,
};

type MessageProps = {
  message: string;
  time: Date;
  type: TaskType;
  status: TaskStatus;
  output: string;
};

export const Message = ({
  time,
  message,
  type,
  status,
  output,
}: MessageProps) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const isInput = type === TaskType.Input;
  const isFailed = status === TaskStatus.Failed;

  const toggleExpand = () => {
    setIsExpanded((prev) => !prev);
  };

  const getMessageStyle = () => {
    if (isInput) return messageStyles.Input;
    return isFailed ? messageStyles.Failed : messageStyles.Regular;
  };

  return (
    <div className={wrapperStyles}>
      <img
        src={isInput ? mePng : logoPng}
        alt={isInput ? "User avatar" : "AI assistant avatar"}
        className={avatarStyles}
        width="40"
        height="40"
      />
      <div className={rightColumnStyles}>
        <div className={timeStyles}>
          {formatDistanceToNowStrict(new Date(time), { addSuffix: true })}
        </div>
        <div className={getMessageStyle()} onClick={toggleExpand}>
          <div className={contentStyles}>
            <span className={clsx(isFailed ? iconStyles.Failed : iconStyles.Regular)}>
              {taskTypeIcons[type]}
            </span>
            <div>{message}</div>
          </div>
          {status === TaskStatus.InProgress && (
            <Button size="small" hierarchy="danger">
              Stop
            </Button>
          )}
        </div>
        {isExpanded && <div className={outputStyles}>{output}</div>}
      </div>
    </div>
  );
};
