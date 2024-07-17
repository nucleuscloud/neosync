import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  refreshLogsWhenRunNotComplete,
  useGetJobRunLogs,
} from '@/libs/hooks/useGetJobRunLogs';
import { LogLevel } from '@neosync/sdk';
import { ReloadIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo, useState } from 'react';
import { getColumns } from './JobRunLogsTable/columns';
import { DataTable } from './JobRunLogsTable/data-table';

interface JobRunLogsProps {
  accountId: string;
  runId: string;
}

export default function JobRunLogs({
  accountId,
  runId,
}: JobRunLogsProps): ReactElement {
  const columns = useMemo(() => getColumns({}), []);
  const [selectedLogLevel, setSelectedLogLevel] = useState<LogLevel>(
    LogLevel.UNSPECIFIED
  );

  const {
    data: logsData,
    isLoading: isLogsLoading,
    isFetching: isLogsValidating,
    refetch: logsMutate,
    error: logsError,
  } = useGetJobRunLogs(runId, accountId, selectedLogLevel, {
    refreshIntervalFn: refreshLogsWhenRunNotComplete,
  });
  const logResponses = logsData ?? [];

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
        columns={columns}
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
