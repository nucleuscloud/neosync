import FastTable from '@/components/FastTable/FastTable';
import { useAccount } from '@/components/providers/account-provider';
import { useQuery } from '@connectrpc/connect-query';
import { JobService, PiiDetectionReport_TableReport } from '@neosync/sdk';
import {
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement, useMemo } from 'react';
import { PII_DETECTION_COLUMNS, PiiDetectionRow } from './columns';

interface Props {
  jobRunId: string;
}

export default function JobRunPiiDetectionTable(props: Props): ReactElement {
  const { jobRunId } = props;
  const { account } = useAccount();
  const _ = jobRunId;
  const {
    data: reportResp,
    isLoading: isLoadingReport,
    isPending: isPendingReport,
  } = useQuery(
    JobService.method.getPiiDetectionReport,
    {
      jobRunId: jobRunId,
      accountId: account?.id,
    },
    {
      enabled: !!account && !!jobRunId,
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
  }, [reportResp?.report, isLoadingReport, isPendingReport]);

  const table = useReactTable({
    data: data,
    columns: PII_DETECTION_COLUMNS,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  return (
    <div>
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
