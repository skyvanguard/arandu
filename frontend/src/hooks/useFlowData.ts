import { useLocalStorage } from "@uidotdev/usehooks";
import { useNavigate, useParams } from "react-router-dom";

import {
  FlowStatus,
  Log,
  Model,
  Task,
  useBrowserUpdatedSubscription,
  useCreateFlowMutation,
  useCreateTaskMutation,
  useFinishFlowMutation,
  useFlowQuery,
  useFlowUpdatedSubscription,
  useTaskAddedSubscription,
  useTerminalLogsAddedSubscription,
} from "@/generated/graphql";

export interface FlowData {
  // Flow identification
  id: string | undefined;
  isNewFlow: boolean;

  // Flow data
  tasks: Task[];
  name: string;
  status: FlowStatus | undefined;
  model: Model | undefined;

  // Terminal data
  terminal: {
    containerName: string | undefined;
    logs: Log[];
    connected: boolean;
  };

  // Browser data
  browser: {
    url: string | undefined;
    screenshotUrl: string;
  };

  // Actions
  handleSubmit: (message: string) => Promise<void>;
  handleFlowStop: () => void;
}

export interface UseFlowDataOptions {
  onBrowserUpdate?: () => void;
  onTerminalUpdate?: () => void;
}

export function useFlowData(options: UseFlowDataOptions = {}): FlowData {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [, createFlowMutation] = useCreateFlowMutation();
  const [, createTaskMutation] = useCreateTaskMutation();
  const [, finishFlowMutation] = useFinishFlowMutation();
  const [selectedModel] = useLocalStorage<Model>("model");

  const isNewFlow = !id || id === "new";

  const [{ operation, data }] = useFlowQuery({
    pause: isNewFlow,
    variables: { id },
  });

  // Handle stale data issue with urql
  // https://github.com/urql-graphql/urql/issues/2507#issuecomment-1159281108
  const isStaleData = operation?.variables.id !== id;

  // Extract flow data with fallbacks
  const flowData = !isStaleData ? data?.flow : undefined;
  const tasks = (flowData?.tasks ?? []) as Task[];
  const name = flowData?.name ?? "";
  const status = flowData?.status;
  const model = flowData?.model as Model | undefined;
  const terminalData = flowData?.terminal;
  const browserData = flowData?.browser;

  // Subscriptions
  useBrowserUpdatedSubscription(
    {
      variables: { flowId: Number(id) },
      pause: isNewFlow,
    },
    () => {
      options.onBrowserUpdate?.();
    }
  );

  useTerminalLogsAddedSubscription(
    {
      variables: { flowId: Number(id) },
      pause: isNewFlow,
    },
    () => {
      options.onTerminalUpdate?.();
    }
  );

  useTaskAddedSubscription({
    variables: { flowId: Number(id) },
    pause: isNewFlow,
  });

  useFlowUpdatedSubscription({
    variables: { flowId: Number(id) },
    pause: isNewFlow,
  });

  // Actions
  const handleSubmit = async (message: string) => {
    if (isNewFlow) {
      if (!selectedModel?.id || !selectedModel?.provider) {
        return;
      }
      const result = await createFlowMutation({
        modelProvider: selectedModel.provider,
        modelId: selectedModel.id,
      });

      const flowId = result?.data?.createFlow.id;
      if (flowId) {
        navigate(`/chat/${flowId}`, { replace: true });
        createTaskMutation({
          flowId: flowId,
          query: message,
        });
      }
    } else if (id) {
      createTaskMutation({
        flowId: id,
        query: message,
      });
    }
  };

  const handleFlowStop = () => {
    if (!id) return;
    finishFlowMutation({ flowId: id });
  };

  return {
    id,
    isNewFlow,
    tasks,
    name,
    status,
    model,
    terminal: {
      containerName: terminalData?.containerName ?? undefined,
      logs: (terminalData?.logs ?? []) as Log[],
      connected: terminalData?.connected ?? false,
    },
    browser: {
      url: browserData?.url ?? undefined,
      screenshotUrl: browserData?.screenshotUrl ?? "",
    },
    handleSubmit,
    handleFlowStop,
  };
}
