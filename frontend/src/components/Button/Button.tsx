import type React from "react";
import clsx from "clsx";
import { forwardRef } from "react";

import {
  buttonIconStyles,
  buttonSizesStyles,
  buttonStyles,
} from "./Button.css";

const hierarchyStyles = {
  primary: buttonStyles.Primary,
  secondary: buttonStyles.Secondary,
  danger: buttonStyles.Danger,
} as const;

const sizeStyles = {
  small: buttonSizesStyles.Small,
  medium: buttonSizesStyles.Medium,
} as const;

export type ButtonProps = {
  children: React.ReactNode;
  icon?: React.ReactNode;
  disabled?: boolean;
  hierarchy?: keyof typeof hierarchyStyles;
  size?: keyof typeof sizeStyles;
} & React.ButtonHTMLAttributes<HTMLButtonElement>;

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      icon = null,
      disabled = false,
      children,
      hierarchy = "primary",
      size = "medium",
      className,
      ...rest
    },
    ref,
  ) => (
    <button
      ref={ref}
      className={clsx(hierarchyStyles[hierarchy], sizeStyles[size], className)}
      disabled={disabled}
      {...rest}
    >
      {icon && <div className={buttonIconStyles}>{icon}</div>}
      {children}
    </button>
  ),
);

Button.displayName = "Button";
