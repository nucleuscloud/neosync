import { SchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import {
  convertJobMappingTransformerToForm,
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import {
  GenerateDefault,
  JobMappingTransformer,
  Passthrough,
  TransformerConfig,
} from '@neosync/sdk';

interface Props {
  getMappings(): JobMappingFormValues[];
  setTransformer(idx: number, transformer: JobMappingTransformerForm): void;
  triggerUpdate(): void;
  constraintHandler: SchemaConstraintHandler;
}

interface UseOnApplyDefaultClickResponse {
  onClick(override: boolean): void;
}

// Hook that provides an onClick handler that will handle setting job mappings to a sensible default transformer
export function useOnApplyDefaultClick(
  props: Props
): UseOnApplyDefaultClickResponse {
  const { getMappings, setTransformer, triggerUpdate, constraintHandler } =
    props;

  return {
    onClick(override) {
      const formMappings = getMappings();
      formMappings.forEach((fm, idx) => {
        // skips setting the default transformer if the user has already set the transformer
        if (fm.transformer.config.case && !override) {
          return;
        } else {
          const colkey = {
            schema: fm.schema,
            table: fm.table,
            column: fm.column,
          };
          const isGenerated = constraintHandler.getIsGenerated(colkey);
          const identityType = constraintHandler.getIdentityType(colkey);
          const newJm = getDefaultJMTransformer(isGenerated && !identityType);
          setTransformer(idx, convertJobMappingTransformerToForm(newJm));
        }
      });
      setTimeout(() => {
        triggerUpdate();
      }, 0);
    },
  };
}

function getDefaultJMTransformer(useDefault: boolean): JobMappingTransformer {
  return useDefault
    ? new JobMappingTransformer({
        config: new TransformerConfig({
          config: {
            case: 'generateDefaultConfig',
            value: new GenerateDefault(),
          },
        }),
      })
    : new JobMappingTransformer({
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough(),
          },
        }),
      });
}
