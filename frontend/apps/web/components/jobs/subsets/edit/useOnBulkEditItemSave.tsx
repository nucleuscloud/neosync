import { SingleSubsetFormValue } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import { SubsetTableRow } from '../SubsetTable/Columns';
import { buildRowKey } from '../utils';

export interface BulkEditItem {
  rowKeys: string[]; // the key of the rows being edited from the tableRowData variable
  item: Pick<SubsetTableRow, 'where'>;
  onClearSelection(): void;
}

interface Props {
  bulkEditItem: BulkEditItem | undefined; // undefined handles the unselected state

  getSubsets(): SingleSubsetFormValue[];
  setSubsets(subsets: SingleSubsetFormValue[]): void;
  appendSubsets(subsets: SingleSubsetFormValue[]): void;
  triggerUpdate(): void;
  getTableRowData(key: string): SubsetTableRow | undefined;
}

interface UseOnBulkEditItemSaveResponse {
  onClick(): void;
}

export default function useOnBulkEditItemSave(
  props: Props
): UseOnBulkEditItemSaveResponse {
  const {
    bulkEditItem,
    getSubsets,
    setSubsets,
    triggerUpdate,
    getTableRowData,
    appendSubsets,
  } = props;

  return {
    onClick() {
      if (!bulkEditItem) {
        return;
      }
      const { rowKeys, onClearSelection, item } = bulkEditItem;
      const subsets = getSubsets();
      const subsetsToEdit = new Map(
        subsets.map((ss) => [buildRowKey(ss.schema, ss.table), ss])
      );
      const subsetsToAdd: SingleSubsetFormValue[] = [];
      rowKeys.forEach((key) => {
        const subset = subsetsToEdit.get(key);
        if (subset) {
          subset.whereClause = item.where;
        } else {
          const td = getTableRowData(key);
          if (td) {
            subsetsToAdd.push({
              schema: td.schema,
              table: td.table,
              whereClause: item.where,
            });
          }
        }
      });
      setSubsets(subsets);
      if (subsetsToAdd.length > 0) {
        appendSubsets(subsetsToAdd);
      }
      setTimeout(() => {
        triggerUpdate();
        onClearSelection();
      }, 0);
    },
  };
}
