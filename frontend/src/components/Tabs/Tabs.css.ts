import { globalStyle, style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const tabsRootStyles = style({
  display: "flex",
  flexDirection: "column",
});

export const tabsListStyles = style({
  flexShrink: 0,
  display: "flex",
  gap: vars.spacing.lg,
  paddingBottom: vars.spacing.xs,
});

export const tabsTriggerStyles = style([
  font.textSmSemibold,
  {
    position: "relative",
    backgroundColor: vars.color.gray2,
    padding: `${vars.spacing.md} ${vars.spacing.sm}`,
    height: vars.spacing["2xl"],
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    userSelect: "none",
    border: "none",
    color: vars.color.gray11,
    borderRadius: vars.radius.lg,
    gap: "6px",

    selectors: {
      '&[data-state="active"]': {
        color: vars.color.gray12,
        backgroundColor: vars.color.gray3,
      },
      '&:hover:not([data-state="active"])': {
        backgroundColor: vars.color.gray3,
        cursor: "pointer",
      },
    },
  },
]);

globalStyle(`${tabsTriggerStyles}:where([data-state="active"]):before`, {
  content: "''",
  position: "absolute",
  left: 0,
  right: 0,
  bottom: -5,
  height: 2,
  backgroundColor: vars.color.primary9,
});

export const tabsContentStyles = style({
  paddingTop: vars.spacing.xl,
});
