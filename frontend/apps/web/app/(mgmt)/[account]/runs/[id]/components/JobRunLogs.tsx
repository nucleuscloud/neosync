import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  refreshLogsWhenRunNotComplete,
  useGetJobRunLogs,
} from '@/libs/hooks/useGetJobRunLogs';
import { ReloadIcon } from '@radix-ui/react-icons';
import { useVirtualizer } from '@tanstack/react-virtual';
import { ReactElement, useRef } from 'react';

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
    refreshIntervalFn: refreshLogsWhenRunNotComplete,
  });
  const logs = logsData || [];

  const parentRef = useRef<HTMLDivElement>(null);

  const count = logs.length;
  const virtualizer = useVirtualizer({
    count,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 45,
  });

  const items = virtualizer.getVirtualItems();

  function onRefreshClick(): void {
    logsMutate();
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
      {logs?.some((l) => l.includes('ERROR')) && (
        <Alert variant="destructive">
          <AlertTitle>{`Log Errors: check logs for errors`}</AlertTitle>
        </Alert>
      )}
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
        <div
          ref={parentRef}
          className="w-100 h-[500px] p-2 border rounded-md dark:border-gray-700 overflow-y-auto	contain-[strict]"
        >
          <div className={`h-[${virtualizer.getTotalSize()}px] w-100 relative`}>
            <div
              className={`absolute w-100 top-0 left-0`}
              style={{
                transform: `translateY(${items[0]?.start ?? 0}px)`, // do not use tailwind for this
              }}
            >
              {items.map((virtualRow) => (
                <div
                  key={virtualRow.key}
                  data-index={virtualRow.index}
                  ref={virtualizer.measureElement}
                >
                  <div className="p-1">
                    <p className="text-sm">{logs[virtualRow.index]}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
