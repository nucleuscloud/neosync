import { ImportMappingsConfig } from '@/components/jobs/SchemaTable/ImportJobMappingsButton';
import {
  convertJobMappingTransformerToForm,
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { JobMapping, JobMappingTransformer } from '@neosync/sdk';

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
        // form.setValue('mappings', [], {
        //   shouldDirty: true,
        //   shouldTouch: true,
        //   shouldValidate: false,
        // });
        // Setting this here in a timeout so that the UI goes through a full render cycle prior to updating the values again
        setTimeout(() => {
          setMappings(formValues);
          // form.setValue('mappings', formValues, {
          //   shouldDirty: true,
          //   shouldTouch: true,
          //   shouldValidate: false,
          // });
        }, 0);
      } else {
        const existingValues = getMappings(); //form.getValues('mappings');
        const existingValueMap: Record<string, number> = {};
        existingValues.forEach((jm, idx) => {
          existingValueMap[`${jm.schema}.${jm.table}.${jm.column}`] = idx;
        });

        const toAdd: JobMappingFormValues[] = [];
        importedJobMappings.forEach((jm) => {
          newSelectedTables.add(`${jm.schema}.${jm.table}`);
          const key = `${jm.schema}.${jm.table}.${jm.column}`;
          const existingIdx = existingValueMap[key] as number | undefined;
          if (existingIdx != null) {
            if (importConfig.overrideOverlapping)
              setTransformer(
                existingIdx,
                convertJobMappingTransformerToForm(
                  jm.transformer ?? new JobMappingTransformer()
                )
              );
            // form.setValue(
            //   `mappings.${existingIdx}.transformer`,
            //   convertJobMappingTransformerToForm(
            //     jm.transformer ?? new JobMappingTransformer()
            //   )
            // );
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
        setSelectedTables(newSelectedTables);
      }
      // wrapping this in a set timeout so the form can go through a render cycle prior to triggering the validation
      setTimeout(() => {
        triggerUpdate();
      }, 0);
    },
  };
}
