import { SingleSubsetFormValue } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import { SubsetTableRow } from '../SubsetTable/Columns';
import { buildRowKey } from '../utils';

interface Props {
  item: SubsetTableRow | undefined; // undefined handles the unselected state

  getSubsets(): SingleSubsetFormValue[];
  appendSubsets(subsets: SingleSubsetFormValue[]): void;
  triggerUpdate(): void;
  updateSubset(idx: number, subset: SingleSubsetFormValue): void;
}

interface UseOnEditItemSaveResponse {
  onClick(): void;
}

export default function useOnEditItemSave(
  props: Props
): UseOnEditItemSaveResponse {
  const { item, getSubsets, triggerUpdate, appendSubsets, updateSubset } =
    props;

  return {
    onClick() {
      if (!item) {
        return;
      }
      const key = buildRowKey(item.schema, item.table);

      const subsets = getSubsets();
      const existingSubsetIdx = subsets.findIndex(
        (ss) => buildRowKey(ss.schema, ss.table) === key
      );
      if (existingSubsetIdx >= 0) {
        updateSubset(existingSubsetIdx, {
          schema: item.schema,
          table: item.table,
          whereClause: item.where,
        });
      } else {
        appendSubsets([
          {
            schema: item.schema,
            table: item.table,
            whereClause: item.where,
          },
        ]);
      }
      setTimeout(() => {
        triggerUpdate();
      }, 0);
    },
  };
}
