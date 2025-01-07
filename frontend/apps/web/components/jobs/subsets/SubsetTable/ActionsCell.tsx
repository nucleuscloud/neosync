import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { Pencil1Icon, ReloadIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';

interface Props {
  onEdit(): void;
  onReset(): void;
  isResetDisabled: boolean;
}

export default function ActionsCell(props: Props): ReactElement {
  const { onEdit, onReset, isResetDisabled } = props;
  return (
    <div className="flex gap-2">
      <EditAction onClick={() => onEdit()} />
      <ResetAction onClick={() => onReset()} isDisabled={isResetDisabled} />
    </div>
  );
}

interface EditActionProps {
  onClick(): void;
}

function EditAction(props: EditActionProps): ReactElement {
  const { onClick } = props;
  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      className="w-8 h-8"
      onClick={() => onClick()}
    >
      <Pencil1Icon />
    </Button>
  );
}

interface ResetActionProps {
  onClick(): void;
  isDisabled: boolean;
}

function ResetAction(props: ResetActionProps): ReactElement {
  const { onClick, isDisabled } = props;
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            type="button"
            className="scale-x-[-1] w-8 h-8"
            variant="outline"
            size="icon"
            onClick={() => onClick()}
            disabled={isDisabled}
          >
            <ReloadIcon />
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>Reset changes made locally to this row</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
