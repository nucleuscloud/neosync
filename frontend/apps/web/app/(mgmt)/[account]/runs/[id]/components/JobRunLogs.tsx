import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { useQuery } from '@connectrpc/connect-query';
import { JobService, LogLevel, LogWindow } from '@neosync/sdk';
import { ReloadIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { getColumns } from './JobRunLogsTable/columns';
import { DataTable } from './JobRunLogsTable/data-table';

interface JobRunLogsProps {
  accountId: string;
  runId: string;
  isRunning: boolean;
}

const TABLE_COLUMNS = getColumns();
const TEN_SECONDS = 5 * 1000;

export default function JobRunLogs({
  accountId,
  runId,
  isRunning,
}: JobRunLogsProps): ReactElement {
  const [selectedLogLevel, setSelectedLogLevel] = useState<LogLevel>(
    LogLevel.UNSPECIFIED
  );

  const {
    data: logsData,
    isLoading: isLogsLoading,
    isFetching: isLogsValidating,
    error: logsError,
    refetch: logsMutate,
  } = useQuery(
    JobService.method.getJobRunLogs,
    {
      accountId,
      jobRunId: runId,
      logLevels: [selectedLogLevel],
      maxLogLines: BigInt(5000),
      window: LogWindow.ONE_DAY,
    },
    {
      enabled: !!runId && !!accountId,
      refetchInterval(query) {
        return query.state.data && isRunning ? TEN_SECONDS : 0;
      },
    }
  );
  const logResponses = logsData?.logLines ?? [];

  function onRefreshClick(): void {
    if (!isLogsValidating) {
      logsMutate();
    }
  }

  return (
    <div className="space-y-4">
      {logsError && (
        <Alert variant="destructive">
          <AlertTitle>{logsError.message}</AlertTitle>
        </Alert>
      )}
      {logResponses?.some((l) => l.logLine.startsWith('[ERROR]')) && (
        <Alert variant="destructive">
          <AlertTitle>{`Log Errors: check logs for errors`}</AlertTitle>
        </Alert>
      )}
      <div className="flex flex-row items-center gap-8">
        <div className="flex flex-row items-center gap-2">
          <h2 className="text-2xl font-bold tracking-tight">Logs</h2>
          <Button
            className={isLogsValidating ? 'animate-spin' : ''}
            disabled={isLogsValidating}
            variant="ghost"
            size="icon"
            onClick={() => onRefreshClick()}
          >
            <ReloadIcon className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <DataTable
        columns={TABLE_COLUMNS}
        data={logResponses}
        getFuzzyFilterValue={(table) =>
          (table.getColumn('logLine')?.getFilterValue() as string) ?? ''
        }
        setFuzzyFilterValue={(table, newval) => {
          table.getColumn('logLine')?.setFilterValue(newval);
        }}
        selectedLogLevel={selectedLogLevel}
        setSelectedLogLevel={setSelectedLogLevel}
        isLoading={isLogsLoading}
      />
    </div>
  );
}
