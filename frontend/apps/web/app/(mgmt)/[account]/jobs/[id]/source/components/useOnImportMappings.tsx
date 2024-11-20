import { ImportMappingsConfig } from '@/components/jobs/SchemaTable/ImportJobMappingsButton';
import {
  convertJobMappingTransformerToForm,
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { JobMapping, JobMappingTransformer } from '@neosync/sdk';
import { toast } from 'sonner';

interface Props {
  setMappings(mappings: JobMappingFormValues[]): void;
  getMappings(): JobMappingFormValues[];
  setTransformer(idx: number, transformer: JobMappingTransformerForm): void;
  appendNewMappings(mappings: JobMappingFormValues[]): void;
  triggerUpdate(): void;

  setSelectedTables(tables: Set<string>): void;
}

interface UseOnImportMappingsResponse {
  onClick(
    importedMappings: JobMapping[],
    importConfig: ImportMappingsConfig
  ): void;
}

// Hook that provides an onClick handler that will import job mappings.
export function useOnImportMappings(props: Props): UseOnImportMappingsResponse {
  const {
    setMappings,
    getMappings,
    setTransformer,
    appendNewMappings,
    triggerUpdate,
    setSelectedTables,
  } = props;

  return {
    onClick(importedJobMappings, importConfig) {
      const newSelectedTables = new Set<string>();
      if (importConfig.truncateAll) {
        const formValues = importedJobMappings.map(
          (jm): JobMappingFormValues => {
            newSelectedTables.add(`${jm.schema}.${jm.table}`);
            return {
              schema: jm.schema,
              table: jm.table,
              column: jm.column,
              transformer: convertJobMappingTransformerToForm(
                jm.transformer ?? new JobMappingTransformer()
              ),
            };
          }
        );
        // Doing this hack here where empty the mappings to force the form to reset IDs.
        // If we don't do this, then the call to form.setValue for mappings will cause a UI bug where
        // any existing rows from the import that may have a different transformer are updated in the form, but not in
        // the UI until a full-refresh takes place (like switching browser tabs and then back again).
        setMappings([]);
        // Setting this here in a timeout so that the UI goes through a full render cycle prior to updating the values again
        setTimeout(() => {
          setMappings(formValues);
          toast.success(`Successfully imported ${formValues.length} mappings!`);
        }, 0);
      } else {
        const existingValues = getMappings();
        const existingValueMap: Record<string, number> = {};
        existingValues.forEach((jm, idx) => {
          existingValueMap[`${jm.schema}.${jm.table}.${jm.column}`] = idx;
        });
        let numReplaced = 0;

        const toAdd: JobMappingFormValues[] = [];
        importedJobMappings.forEach((jm) => {
          newSelectedTables.add(`${jm.schema}.${jm.table}`);
          const key = `${jm.schema}.${jm.table}.${jm.column}`;
          const existingIdx = existingValueMap[key] as number | undefined;
          if (existingIdx != null) {
            if (importConfig.overrideOverlapping) {
              numReplaced += 1;
              setTransformer(
                existingIdx,
                convertJobMappingTransformerToForm(
                  jm.transformer ?? new JobMappingTransformer()
                )
              );
            }
          } else {
            toAdd.push({
              schema: jm.schema,
              table: jm.table,
              column: jm.column,
              transformer: convertJobMappingTransformerToForm(
                jm.transformer ?? new JobMappingTransformer()
              ),
            });
          }
        });
        if (toAdd.length > 0) {
          appendNewMappings(toAdd);
        }
        toast.success(buildOverlapMessage(numReplaced, toAdd.length));
      }
      setSelectedTables(newSelectedTables);
      // wrapping this in a set timeout so the form can go through a render cycle prior to triggering the validation
      setTimeout(() => {
        triggerUpdate();
      }, 0);
    },
  };
}

function buildOverlapMessage(numReplaced: number, numAdded: number): string {
  if (numReplaced === 0 && numAdded === 0) {
    return 'Import succeeded but did not add or replace any new mappings.';
  }

  return `Import succeeded with ${numReplaced || 'no'} mappings replaced and ${numAdded || 'none'} added.`;
}
