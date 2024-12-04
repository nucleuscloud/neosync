import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { useQuery } from '@connectrpc/connect-query';
import { JobHook, JobHookConfig, JobHookConfig_JobSqlHook } from '@neosync/sdk';
import { getConnections } from '@neosync/sdk/connectquery';
import { ClockIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import EditHookButton from './EditHookButton';
import RemoveHookButton from './RemoveHookButton';

interface Props {
  hook: JobHook;
  jobConnectionIds: string[];
  onDeleted(): void;
  onEdited(): void;
}

export default function HookCard(props: Props): ReactElement {
  const { hook, onDeleted, onEdited, jobConnectionIds } = props;
  const hookTiming = getHookTiming(hook.config ?? new JobHookConfig());

  return (
    <div id={`jobhook-${hook.id}`}>
      <Card>
        <CardContent className="p-4">
          <div className="flex justify-between">
            <div className="flex flex-grow flex-col gap-2">
              {/* Header Section */}
              <div className="flex items-center justify-between">
                <h3 className="font-medium text-lg">{hook.name}</h3>
              </div>

              <p className="text-sm text-gray-600 dark:text-gray-400">
                {hook.description}
              </p>

              <div className="flex items-center gap-4 pt-2 flex-wrap">
                <div className="flex items-center gap-1 text-sm text-gray-600 dark:text-gray-400">
                  <ClockIcon className="flex-shrink-0" />
                  <p>{hookTiming}</p>
                </div>

                <Badge variant="secondary">
                  <span>P</span>
                  {hook.priority}
                </Badge>
                <Badge variant={hook.enabled ? 'success' : 'outline'}>
                  {hook.enabled ? 'Enabled' : 'Disabled'}
                </Badge>
                {hook.config?.config.case && (
                  <Badge variant="secondary">{hook.config?.config.case}</Badge>
                )}
                {hook.config?.config.case && (
                  <HookConnectionBadge config={hook.config} />
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <EditHookButton
                hook={hook}
                onEdited={onEdited}
                jobConnectionIds={jobConnectionIds}
              />
              <RemoveHookButton hook={hook} onDeleted={onDeleted} />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function getHookTiming(config: JobHookConfig): string | undefined {
  switch (config.config.case) {
    case 'sql': {
      switch (config.config.value.timing?.timing.case) {
        case 'preSync': {
          return 'Pre Sync';
        }
        case 'postSync': {
          return 'Post Sync';
        }
        default: {
          return config.config.value.timing?.timing.case;
        }
      }
    }
  }
  return 'Unknown';
}

interface HookConnectionBadgeProps {
  config: JobHookConfig;
}
function HookConnectionBadge(
  props: HookConnectionBadgeProps
): ReactElement | null {
  const { config } = props;

  switch (config.config.case) {
    case 'sql': {
      return <SqlHookConnectionBadge config={config.config.value} />;
    }
    default: {
      return null;
    }
  }
}

interface SqlHookConnectionBadgeProps {
  config: Pick<JobHookConfig_JobSqlHook, 'connectionId'>;
}
function SqlHookConnectionBadge(
  props: SqlHookConnectionBadgeProps
): ReactElement {
  const { config } = props;

  const { account } = useAccount();
  const { data: getConnectionsResp } = useQuery(
    getConnections, // Using getConnections here so that way all of the badges can share the same request
    { accountId: account?.id },
    { enabled: !!account?.id }
  );

  const connection = getConnectionsResp?.connections.find(
    (c) => c.id === config.connectionId
  );

  return (
    <Badge variant="secondary">
      {connection?.name ?? 'Unknown Connection'}
    </Badge>
  );
}

function getHookConnectionName(
  config: JobHookConfig,
  getConnectionName: (id: string) => string
): string | undefined {
  switch (config.config.case) {
    case 'sql': {
      return getConnectionName(config.config.value.connectionId);
    }
  }
  return '';
}
