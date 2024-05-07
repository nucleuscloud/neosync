import { ColumnDef } from '@tanstack/react-table';
import { SampleRecord } from '../types';

export function getAiSampleTableColumns(
  keys: string[]
): ColumnDef<SampleRecord>[] {
  return keys.map((key) => {
    return {
      accessorKey: key,
      header: () => <p>{key}</p>,
      cell: ({ getValue }) => (
        <div>
          <p>{getValue<string>()}</p>
        </div>
      ),
    };
  });
}
