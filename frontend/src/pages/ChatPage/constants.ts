// Tab identifiers
export const TAB_IDS = {
  TERMINAL: "terminal",
  BROWSER: "browser",
  CODE: "code",
} as const;

export type TabId = (typeof TAB_IDS)[keyof typeof TAB_IDS];

// Default tab
export const DEFAULT_TAB: TabId = TAB_IDS.TERMINAL;

// Local storage keys
export const STORAGE_KEYS = {
  IS_FOLLOWING_TABS: "isFollowingTabs",
} as const;
