import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  refreshLogsWhenRunNotComplete,
  useGetJobRunLogs,
} from '@/libs/hooks/useGetJobRunLogs';
import { ReloadIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';
import { VariableSizeList as List } from 'react-window';

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

  const [windowWidth, setWindowWidth] = useState<number>(window.innerWidth);
  const listRef = useRef<List<string[]> | null>(null);

  useEffect(() => {
    function handleResize() {
      setWindowWidth(window.innerWidth);
      if (listRef.current) {
        listRef.current.resetAfterIndex(0);
      }
    }

    window.addEventListener('resize', handleResize);
    handleResize();
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  function onRefreshClick(): void {
    logsMutate();
  }

  function getLogLineSize(index: number): number {
    const log = logs[index];
    const maxLineWidth = windowWidth;
    const estimatedLineWidth = log.length * 10;
    const numberOfLines = Math.ceil(estimatedLineWidth / maxLineWidth) + 1;
    const height = 5 + numberOfLines * 20;
    return height;
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
        <div className="h-[500px] w-full">
          <AutoSizer>
            {({ height, width }) => (
              <List
                ref={listRef}
                className="border rounded-md dark:border-gray-700"
                height={height}
                itemCount={logs.length}
                itemSize={getLogLineSize}
                width={width}
                itemKey={(index: number) => logs[index]}
                itemData={logs}
              >
                {({ index, style }) => {
                  return (
                    <p className="p-2 text-sm" style={style}>
                      {logs[index]}
                    </p>
                  );
                }}
              </List>
            )}
          </AutoSizer>
        </div>
      )}
    </div>
  );
}
