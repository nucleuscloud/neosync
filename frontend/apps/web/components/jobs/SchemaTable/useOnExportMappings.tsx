import {
  convertJobMappingTransformerFormToJobMappingTransformer,
  JobMappingFormValues,
} from '@/yup-validations/jobs';
import { create, toJson } from '@bufbuild/protobuf';
import { JobMappingSchema } from '@neosync/sdk';
import { Row } from '@tanstack/react-table';
import { toast } from 'sonner';
import { useJsonFileDownload } from '../../useJsonFileDownload';

interface Props {
  jobMappings: JobMappingFormValues[];
}

interface UseOnExportMappingsResponse<T> {
  onClick(selectedRows: Row<T>[], shouldFormat: boolean): Promise<void>;
}

// Hook that provides an onClick handler that will download job mappings to disk
export function useOnExportMappings<T>(
  props: Props
): UseOnExportMappingsResponse<T> {
  const { jobMappings } = props;
  const { downloadFile } = useJsonFileDownload();

  return {
    onClick: async function (
      selectedRows: Row<T>[],
      shouldFormat: boolean
    ): Promise<void> {
      // Using the raw jobMappings instead of the row due to tanstack sometimes not giving the most up to date values.
      const dataToDownload =
        selectedRows.length === 0
          ? jobMappings
          : selectedRows.map((row) => jobMappings[row.index]);
      const jms = dataToDownload.map((d) => {
        return toJson(
          JobMappingSchema,
          create(JobMappingSchema, {
            schema: d.schema,
            table: d.table,
            column: d.column,
            transformer:
              convertJobMappingTransformerFormToJobMappingTransformer(
                d.transformer
              ),
          }),
          {
            // Forces the proto field names to be snake_case instead of lowerCamelCase.
            // This is because the ES format changes it to be lowerCamelCase, but this means that the mappings
            // can't be used natively by any of the other SDKs.
            // Using this format makes it uniformly accepted by all of our SDKs.
            // This does not affect the import as it can natively handle both scenarios.
            useProtoFieldName: true,
          }
        );
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
