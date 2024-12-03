import {
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';

interface Props {
  getMappings(): JobMappingFormValues[];
  setMappings(mappings: JobMappingFormValues[]): void;
  triggerUpdate(): void;
}

interface UseOnTransformerBulkUpdateClickResponse {
  onClick(indices: number[], transformer: JobMappingTransformerForm): void;
}

// Hook that provides an onClick handler that will handle setting specified job mappings to a the provided transformer
export function useOnTransformerBulkUpdateClick(
  props: Props
): UseOnTransformerBulkUpdateClickResponse {
  const { getMappings, triggerUpdate, setMappings } = props;

  return {
    onClick(indices, transformer) {
      if (indices.length === 0) {
        return;
      }

      const formMappings = getMappings();

      indices.forEach((idx) => {
        if (formMappings[idx]) {
          formMappings[idx].transformer = transformer;
        }
      });

      setMappings(formMappings);
      setTimeout(() => {
        triggerUpdate();
      }, 0);
    },
  };
}
