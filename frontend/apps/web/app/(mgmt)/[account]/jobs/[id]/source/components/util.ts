import { Action } from '@/components/DualListBox/DualListBox';
import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import {
  JobMappingFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { JobMappingTransformer, JobSource } from '@neosync/sdk';

export function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3'
  ) {
    return js.options.config.value.connectionId;
  }
  return undefined;
}

export function getFkIdFromGenerateSource(
  js: JobSource | undefined
): string | undefined {
  if (js?.options?.config.case === 'generate') {
    return js.options.config.value.fkSourceConnectionId;
  }
  if (js?.options?.config.case === 'aiGenerate') {
    return js.options.config.value.fkSourceConnectionId;
  }
  return undefined;
}

function getSetDelta(
  newSet: Set<string>,
  oldSet: Set<string>
): [Set<string>, Set<string>] {
  const added = new Set<string>();
  const removed = new Set<string>();

  oldSet.forEach((val) => {
    if (!newSet.has(val)) {
      removed.add(val);
    }
  });
  newSet.forEach((val) => {
    if (!oldSet.has(val)) {
      added.add(val);
    }
  });

  return [added, removed];
}

export function getOnSelectedTableToggle(
  schema: ConnectionSchemaMap,
  selectedTables: Set<string>,
  setSelectedTables: (newitems: Set<string>) => void,
  fields: { schema: string; table: string }[],
  remove: (indices: number[]) => void,
  append: (formValues: JobMappingFormValues[]) => void
): (items: Set<string>, action: Action) => void {
  return (items) => {
    if (items.size === 0) {
      const idxs = fields.map((_, idx) => idx);
      remove(idxs);
      setSelectedTables(new Set());
      return;
    }
    const [added, removed] = getSetDelta(items, selectedTables);
    const toRemove: number[] = [];
    const toAdd: JobMappingFormValues[] = [];
    fields.forEach((field, idx) => {
      if (removed.has(`${field.schema}.${field.table}`)) {
        toRemove.push(idx);
      }
    });

    added.forEach((item) => {
      const dbcols = schema[item];
      if (!dbcols) {
        return;
      }
      dbcols.forEach((dbcol) => {
        toAdd.push({
          schema: dbcol.schema,
          table: dbcol.table,
          column: dbcol.column,
          transformer: convertJobMappingTransformerToForm(
            new JobMappingTransformer({})
          ),
        });
      });
    });
    if (toRemove.length > 0) {
      remove(toRemove);
    }
    if (toAdd.length > 0) {
      append(toAdd);
    }
    setSelectedTables(items);
  };
}
