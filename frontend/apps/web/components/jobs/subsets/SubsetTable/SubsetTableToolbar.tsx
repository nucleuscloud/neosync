import ButtonText from '@/components/ButtonText';
import { Button } from '@/components/ui/button';
import { Cross2Icon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';

interface Props {
  isFilterButtonDisabled: boolean;
  onClearFilters(): void;

  isBulkEditButtonDisabled: boolean;
  onBulkEditClick(): void;
}
export function SubsetTableToolbar(props: Props): ReactElement {
  const {
    onClearFilters,
    isFilterButtonDisabled,
    isBulkEditButtonDisabled,
    onBulkEditClick,
  } = props;

  return (
    <div className="flex flex-row items-center gap-2 justify-end">
      <div className="flex flex-row items-center gap-2">
        <Button
          disabled={isBulkEditButtonDisabled}
          type="button"
          variant="outline"
          className="px-2 lg:px-3"
          onClick={onBulkEditClick}
        >
          <ButtonText text="Bulk Edit Subsets" />
        </Button>
        <Button
          disabled={isFilterButtonDisabled}
          type="button"
          variant="outline"
          className="px-2 lg:px-3"
          onClick={onClearFilters}
        >
          <ButtonText
            leftIcon={<Cross2Icon className="h-3 w-3" />}
            text="Clear filters"
          />
        </Button>
      </div>
    </div>
  );
}
