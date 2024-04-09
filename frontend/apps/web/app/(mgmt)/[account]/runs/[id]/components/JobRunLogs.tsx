import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { useGetJobRunLogs } from '@/libs/hooks/useGetJobRunLogs';
import { ReloadIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo } from 'react';
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
  const {
    data: logsData,
    isLoading: isLogsLoading,
    isValidating: isLogsValidating,
    mutate: logsMutate,
    error: logsError,
  } = useGetJobRunLogs(runId, accountId, {
    // refreshIntervalFn: refreshLogsWhenRunNotComplete,
  });
  const logResponses = logsData ?? [];
  const columns = useMemo(() => getColumns({}), []);

  function onRefreshClick(): void {
    if (!isLogsValidating) {
      logsMutate();
    }
  }

  if (logsError) {
    return (
      <Alert variant="destructive">
        <AlertTitle>{logsError.message}</AlertTitle>
      </Alert>
    );
  }

  return (
    <div className="space-y-4">
      {logResponses?.some((l) => l.logLine.startsWith('[ERROR]')) && (
        <Alert variant="destructive">
          <AlertTitle>{`Log Errors: check logs for errors`}</AlertTitle>
        </Alert>
      )} */}
      <div className="flex flex-row items-center space-x-2">
        <h1 className="text-2xl font-bold tracking-tight">Logs</h1>
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
      {isLogsLoading ? (
        <SkeletonTable />
      ) : (
        <DataTable columns={columns} data={logResponses} />
      )}
    </div>
  );
}
