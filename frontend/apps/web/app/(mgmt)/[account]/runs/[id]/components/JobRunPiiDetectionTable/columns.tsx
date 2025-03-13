import { SchemaColumnHeader } from '@/components/jobs/SchemaTable/SchemaColumnHeader';
import TruncatedText from '@/components/TruncatedText';
import { ColumnDef, createColumnHelper } from '@tanstack/react-table';
import CategoryCell from './CategoryCell';
import ConfidenceCell from './ConfidenceCell';
import ReporterTypeCell from './ReporterTypeCell';

// interface Reporter {
//   name: string;
//   confidence?: number;
//   category: string;
// }
export interface PiiDetectionRow {
  schema: string;
  table: string;
  column: string;
  reporterType: string[];
  reporterConfidence: number[];
  reporterCategory: string[];
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getPiiDetectionColumns(): ColumnDef<PiiDetectionRow, any>[] {
  const columnHelper = createColumnHelper<PiiDetectionRow>();

  const schemaColumn = columnHelper.accessor('schema', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Schema" />;
    },
    cell({ getValue }) {
      return <TruncatedText text={getValue()} />;
    },
  });

  const tableColumn = columnHelper.accessor('table', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Table" />;
    },
    cell({ getValue }) {
      return <TruncatedText text={getValue()} />;
    },
  });

  const columnColumn = columnHelper.accessor('column', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Column" />;
    },
    cell({ getValue }) {
      return <TruncatedText text={getValue()} />;
    },
  });

  const reporterTypeColumn = columnHelper.accessor(
    (row) => {
      return row.reporterType.join(', ');
    },
    {
      id: 'reporterType',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Reporter" />;
      },
      cell({ row }) {
        return <ReporterTypeCell reporterTypes={row.original.reporterType} />;
      },
    }
  );

  const reporterConfidenceColumn = columnHelper.accessor(
    (row) => {
      return row.reporterConfidence.join(', ');
    },
    {
      id: 'reporterConfidence',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Confidence" />;
      },
      cell({ row }) {
        return <ConfidenceCell confidence={row.original.reporterConfidence} />;
      },
    }
  );

  const reporterCategoryColumn = columnHelper.accessor(
    (row) => {
      return row.reporterCategory.join(', ');
    },
    {
      id: 'reporterCategory',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Category" />;
      },
      cell({ row }) {
        return <CategoryCell categories={row.original.reporterCategory} />;
      },
    }
  );

  return [
    schemaColumn,
    tableColumn,
    columnColumn,
    reporterTypeColumn,
    reporterConfidenceColumn,
    reporterCategoryColumn,
  ];
}

export const PII_DETECTION_COLUMNS = getPiiDetectionColumns();
