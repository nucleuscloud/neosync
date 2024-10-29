import {
  convertJobMappingTransformerFormToJobMappingTransformer,
  JobMappingFormValues,
} from '@/yup-validations/jobs';
import { JobMapping } from '@neosync/sdk';
import { Row } from '@tanstack/react-table';
import { toast } from 'sonner';
import { useJsonFileDownload } from '../../useJsonFileDownload';

interface Props {
  jobMappings: JobMappingFormValues[];
}

interface UseOnExportMappingsResponse {
  onClick(
    selectedRows: Row<JobMappingFormValues>[],
    shouldFormat: boolean
  ): Promise<void>;
}

// Hook that provides an onClick handler that will download job mappings to disk
export function useOnExportMappings(props: Props): UseOnExportMappingsResponse {
  const { jobMappings } = props;
  const { downloadFile } = useJsonFileDownload();

  return {
    onClick: async function (
      selectedRows: Row<JobMappingFormValues>[],
      shouldFormat: boolean
    ): Promise<void> {
      // Using the raw jobMappings instead of the row due to tanstack sometimes not giving the most up to date values.
      const dataToDownload =
        selectedRows.length === 0
          ? jobMappings
          : selectedRows.map((row) => jobMappings[row.index]);
      const jms = dataToDownload.map((d) => {
        return new JobMapping({
          schema: d.schema,
          table: d.table,
          column: d.column,
          transformer: convertJobMappingTransformerFormToJobMappingTransformer(
            d.transformer
          ),
        }).toJson();
      });
      await downloadFile({
        data: jms,
        fileName: `mappings-${new Date().toISOString()}.json`,
        shouldFormat,
        // todo: may want to export these in the Props in the future for more control
        onSuccess() {
          toast.success('Successfully exported job mappings!');
        },
        onError(error) {
          toast.error(`Failed to export job mappings: ${error}`);
        },
      });
    },
  };
}
