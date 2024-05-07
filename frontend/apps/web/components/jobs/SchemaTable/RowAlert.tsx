import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { Row } from '@tanstack/react-table';
import { ReactElement } from 'react';
import { SchemaConstraintHandler } from './schema-constraint-handler';

interface Props {
  row: Row<{ schema: string; table: string; column: string }>;
  handler: SchemaConstraintHandler;
  onRemoveClick(): void;
}

export default function SchemaRowAlert(props: Props): ReactElement {
  const { row, handler, onRemoveClick } = props;
  const key = {
    schema: row.getValue<string>('schema'),
    table: row.getValue<string>('table'),
    column: row.getValue<string>('column'),
  };
  const isInSchema = handler.getIsInSchema(key);

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
