import FastTable from '@/components/FastTable/FastTable';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { refreshWhenJobRunning } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { JobService, PiiDetectionReport_TableReport } from '@neosync/sdk';
import { DownloadIcon, ReloadIcon } from '@radix-ui/react-icons';
import {
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import Papa from 'papaparse';
import { ReactElement, useMemo } from 'react';
import { PII_DETECTION_COLUMNS, PiiDetectionRow } from './columns';

interface Props {
  jobRunId: string;
  isRunning: boolean;
}

export default function JobRunPiiDetectionTable(props: Props): ReactElement {
  const { jobRunId, isRunning } = props;
  const { account } = useAccount();

  const {
    data: reportResp,
    isLoading: isLoadingReport,
    isPending: isPendingReport,
    isFetching: isFetchingReport,
    refetch: reportMutate,
  } = useQuery(
    JobService.method.getPiiDetectionReport,
    {
      jobRunId: jobRunId,
      accountId: account?.id,
    },
    {
      enabled: !!account && !!jobRunId,
      refetchInterval(query) {
        return query.state.data ? refreshWhenJobRunning(isRunning) : 0;
      },
    }
  );

  const data: PiiDetectionRow[] = useMemo(() => {
    if (!reportResp?.report || isLoadingReport || isPendingReport) {
      return [];
    }
    const report = reportResp.report;
    if (!report) {
      return [];
    }
    return getPiiDetectionRowsFromTables(report.tables);
  }, [reportResp?.report, isLoadingReport, isPendingReport, isFetchingReport]);

  const table = useReactTable({
    data: data,
    columns: PII_DETECTION_COLUMNS,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  function onRefreshClick(): void {
    if (!isFetchingReport) {
      reportMutate();
    }
  }

  function onDownloadClick(): void {
    if (!data.length) {
      return;
    }

    // Create CSV content
    const csvRows = data.map((row) => {
      return {
        schema: row.schema,
        table: row.table,
        column: row.column,
        categories: row.reporterCategory,
        confidence: row.reporterConfidence,
        reporters: row.reporterType,
      };
    });
    const csvContent = Papa.unparse(csvRows, {
      escapeFormulae: true,
      header: true,
    });

    // Create and download the CSV file
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', `pii-detection-report-${jobRunId}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-row gap-2 items-center">
        <div className="text-xl font-semibold tracking-tight">
          PII Detection Report
        </div>
        <div className="flex flex-row gap-1">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onRefreshClick()}
            className={isFetchingReport ? 'animate-spin' : ''}
          >
            <ReloadIcon className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onDownloadClick()}
            disabled={!data.length}
          >
            <DownloadIcon className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <FastTable table={table} estimateRowSize={() => 53} rowOverscan={50} />
    </div>
  );
}

function getPiiDetectionRowsFromTables(
  tables: PiiDetectionReport_TableReport[]
): PiiDetectionRow[] {
  return tables.flatMap(tableReportToPiiDetectionRow);
}

function tableReportToPiiDetectionRow(
  report: PiiDetectionReport_TableReport
): PiiDetectionRow[] {
  const output: PiiDetectionRow[] = [];

  for (const column of report.columns) {
    const categories: string[] = [];
    const confidence: number[] = [];
    const reporters: string[] = [];
    if (column.regexReport) {
      categories.push(column.regexReport.category);
      reporters.push('regex');
    }
    if (column.llmReport) {
      categories.push(column.llmReport.category);
      reporters.push('llm');
      confidence.push(column.llmReport.confidence);
    }
    output.push({
      schema: report.schema,
      table: report.table,
      column: column.column,
      reporterCategory: categories,
      reporterConfidence: confidence,
      reporterType: reporters,
    });
  }

  return output;
}
