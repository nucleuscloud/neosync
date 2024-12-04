import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import {
  Connection,
  JobHook,
  JobHookConfig,
  JobHookConfig_JobSqlHook,
} from '@neosync/sdk';
import { ClockIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo } from 'react';
import EditHookButton from './EditHookButton';
import RemoveHookButton from './RemoveHookButton';

interface Props {
  hook: JobHook;
  jobConnections: Connection[];
  onDeleted(): void;
  onEdited(): void;
}

export default function HookCard(props: Props): ReactElement {
  const { hook, onDeleted, onEdited, jobConnections } = props;
  const hookTiming = getHookTiming(hook.config ?? new JobHookConfig());

  const connectionMap = useMemo(
    () => new Map(jobConnections.map((conn) => [conn.id, conn])),
    [jobConnections]
  );

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
                  <HookConnectionBadge
                    config={hook.config}
                    connMap={connectionMap}
                  />
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <EditHookButton
                hook={hook}
                onEdited={onEdited}
                jobConnections={jobConnections}
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
  connMap: Map<string, Connection>;
}
function HookConnectionBadge(
  props: HookConnectionBadgeProps
): ReactElement | null {
  const { config, connMap } = props;

  switch (config.config.case) {
    case 'sql': {
      return (
        <SqlHookConnectionBadge
          config={config.config.value}
          connMap={connMap}
        />
      );
    }
    default: {
      return null;
    }
  }
}

interface SqlHookConnectionBadgeProps {
  config: Pick<JobHookConfig_JobSqlHook, 'connectionId'>;
  connMap: Map<string, Connection>;
}
function SqlHookConnectionBadge(
  props: SqlHookConnectionBadgeProps
): ReactElement {
  const { config, connMap } = props;
  const connection = connMap.get(config.connectionId);
  return (
    <Badge variant="secondary">
      {connection?.name ?? 'Unknown Connection'}
    </Badge>
  );
}
