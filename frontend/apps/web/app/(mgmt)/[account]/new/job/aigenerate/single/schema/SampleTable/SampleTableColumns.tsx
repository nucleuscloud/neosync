import { ColumnDef } from '@tanstack/react-table';
import { SampleRecord } from '../types';

export function getAiSampleTableColumns(
  keys: string[]
): ColumnDef<SampleRecord>[] {
  return keys.map((key) => {
    return {
      accessorKey: key,
      header: () => <p>{key}</p>,
      cell: ({ getValue }) => {
        const rawValue = getValue();
        const value =
          typeof rawValue === 'string' ? rawValue : JSON.stringify(rawValue);
        return (
          <div>
            <p>{value}</p>
          </div>
        );
      },
    };
  });
}
