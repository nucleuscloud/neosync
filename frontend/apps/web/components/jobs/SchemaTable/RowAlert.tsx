import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import {
  ColumnKey,
  SchemaConstraintHandler,
} from './schema-constraint-handler';

interface Props {
  rowKey: ColumnKey;
  handler: SchemaConstraintHandler;
  onRemoveClick(): void;
}

export default function SchemaRowAlert(props: Props): ReactElement {
  const { rowKey, handler, onRemoveClick } = props;
  const isInSchema = handler.getIsInSchema(rowKey);

  const messages: string[] = [];

  if (!isInSchema) {
    messages.push('This column was not found in the backing source schema');
  }

  if (messages.length === 0) {
    return <div className="hidden" />;
  }

  return (
    <TooltipProvider delayDuration={100}>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="cursor-default">
            <ExclamationTriangleIcon
              className="text-yellow-600 dark:text-yellow-300 cursor-pointer"
              onClick={() => onRemoveClick()}
            />
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>{messages.join('\n')}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
