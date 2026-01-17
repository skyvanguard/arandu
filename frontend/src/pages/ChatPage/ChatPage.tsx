import * as Tabs from "@radix-ui/react-tabs";
import { useLocalStorage } from "@uidotdev/usehooks";
import { useState } from "react";

import Browser from "@/components/Browser/Browser";
import { Button } from "@/components/Button/Button";
import { Icon } from "@/components/Icon/Icon";
import { Messages } from "@/components/Messages/Messages";
import { Panel } from "@/components/Panel/Panel";
import {
  tabsContentStyles,
  tabsListStyles,
  tabsRootStyles,
  tabsTriggerStyles,
} from "@/components/Tabs/Tabs.css";
import { Terminal } from "@/components/Terminal/Terminal";
import { Tooltip } from "@/components/Tooltip/Tooltip";
import { useFlowData } from "@/hooks/useFlowData";

import {
  followButtonStyles,
  leftColumnStyles,
  tabsStyles,
  wrapperStyles,
} from "./ChatPage.css";
import { DEFAULT_TAB, STORAGE_KEYS, TAB_IDS, TabId } from "./constants";

export const ChatPage = () => {
  const [isFollowingTabs, setIsFollowingTabs] = useLocalStorage(
    STORAGE_KEYS.IS_FOLLOWING_TABS,
    true,
  );
  const [activeTab, setActiveTab] = useState<TabId>(DEFAULT_TAB);

  const flow = useFlowData({
    onBrowserUpdate: () => {
      if (isFollowingTabs) setActiveTab(TAB_IDS.BROWSER);
    },
    onTerminalUpdate: () => {
      if (isFollowingTabs) setActiveTab(TAB_IDS.TERMINAL);
    },
  });

  const handleChangeIsFollowingTabs = () => {
    setIsFollowingTabs(!isFollowingTabs);
  };

  const handleTabChange = (value: string) => {
    setActiveTab(value as TabId);
  };

  return (
    <div className={wrapperStyles}>
      <Panel>
        <Messages
          tasks={flow.tasks}
          name={flow.name}
          onSubmit={flow.handleSubmit}
          flowStatus={flow.status}
          isNew={flow.isNewFlow}
          onFlowStop={flow.handleFlowStop}
          model={flow.model}
        />
      </Panel>
      <Panel>
        <Tabs.Root
          className={tabsRootStyles}
          value={activeTab}
          onValueChange={handleTabChange}
        >
          <Tabs.List className={tabsListStyles}>
            <div className={tabsStyles}>
              <div className={leftColumnStyles}>
                <Tabs.Trigger className={tabsTriggerStyles} value={TAB_IDS.TERMINAL}>
                  Terminal
                </Tabs.Trigger>
                <Tabs.Trigger className={tabsTriggerStyles} value={TAB_IDS.BROWSER}>
                  Browser
                </Tabs.Trigger>
                <Tabs.Trigger
                  className={tabsTriggerStyles}
                  value={TAB_IDS.CODE}
                  disabled
                >
                  Code (Soon)
                </Tabs.Trigger>
              </div>

              <Tooltip
                content={
                  <>
                    Following the active tab is{" "}
                    <b>{isFollowingTabs ? "enabled" : "disabled"}</b>
                  </>
                }
              >
                <Button
                  size="small"
                  hierarchy={isFollowingTabs ? "primary" : "secondary"}
                  className={followButtonStyles}
                  onClick={handleChangeIsFollowingTabs}
                >
                  {isFollowingTabs ? <Icon.Eye /> : <Icon.EyeOff />}
                </Button>
              </Tooltip>
            </div>
          </Tabs.List>
          <Tabs.Content className={tabsContentStyles} value={TAB_IDS.TERMINAL}>
            <Terminal
              id={flow.isNewFlow ? "" : flow.id}
              status={flow.status}
              title={flow.terminal.containerName}
              logs={flow.terminal.logs}
              isRunning={flow.terminal.connected}
            />
          </Tabs.Content>
          <Tabs.Content className={tabsContentStyles} value={TAB_IDS.BROWSER}>
            <Browser
              url={flow.browser.url}
              screenshotUrl={flow.browser.screenshotUrl}
            />
          </Tabs.Content>
          <Tabs.Content className={tabsContentStyles} value={TAB_IDS.CODE}>
            code
          </Tabs.Content>
        </Tabs.Root>
      </Panel>
    </div>
  );
};
