import ButtonText from '@/components/ButtonText';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ReactElement } from 'react';
import { SubsetTableRow } from '../SubsetTable/Columns';
import WhereEditor from './WhereEditor';

interface Props {
  item: Pick<SubsetTableRow, 'where'>;
  onItem(item: Pick<SubsetTableRow, 'where'>): void;
  onSave(): void;
  onCancel(): void;
  columns: string[];
}

export default function EditItems(props: Props): ReactElement {
  const { item, onItem, onSave, onCancel, columns } = props;

  function onWhereChange(whereClause: string): void {
    onItem({ where: whereClause });
  }

  return (
    <div className="flex flex-col gap-4">
      <WhereEditor
        whereClause={item?.where ?? ''}
        onWhereChange={onWhereChange}
        columns={columns}
      />
      <div className="flex justify-between gap-4">
        <Button
          type="button"
          variant="secondary"
          disabled={!item}
          onClick={onCancel}
        >
          <ButtonText text="Cancel" />
        </Button>
        <div className="flex flex-row gap-2">
          <TooltipProvider>
            <Tooltip delayDuration={200}>
              <TooltipTrigger asChild>
                <Button type="button" disabled={!item} onClick={() => onSave()}>
                  <ButtonText text="Apply" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>
                  Applies changes to table only, click Save below to fully
                  submit changes
                </p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>
    </div>
  );
}
