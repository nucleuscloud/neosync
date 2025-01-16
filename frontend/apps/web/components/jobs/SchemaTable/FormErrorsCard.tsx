import Spinner from '@/components/Spinner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  CheckCircledIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  ReloadIcon,
} from '@radix-ui/react-icons';
import { ReactElement } from 'react';

export type ErrorLevel = 'error' | 'warning';

export interface FormError {
  message: string;
  type?: string;
  path: string;
  level: ErrorLevel;
}

interface Props {
  formErrors: FormError[];
  isValidating?: boolean;
  onValidate?(): void;
}

export default function FormErrorsCard(props: Props): ReactElement {
  const { formErrors, isValidating, onValidate } = props;

  const messages = formErrorsToMessages(formErrors);
  const warnings = formWarningsToMessages(formErrors);
  return (
    <Card className="w-full flex flex-col">
      <CardHeader className="flex flex-col">
        <div className="flex flex-row items-center justify-between h-8">
          <div className="flex flex-row items-center gap-2">
            {messages.length != 0 ? (
              <ExclamationTriangleIcon className="h-4 w-4 text-destructive dark:text-red-400 text-red-600" />
            ) : (
              <CheckCircledIcon className="w-4 h-4" />
            )}
            <CardTitle>Validations</CardTitle>
            {messages.length != 0 && (
              <Badge variant="destructive">
                {messages.length == 1
                  ? `${messages.length} Error`
                  : `${messages.length} Errors`}
              </Badge>
            )}
            {warnings.length != 0 && (
              <Badge className="bg-yellow-200 dark:bg-yellow-800/70 text-yellow-900 dark:text-yellow-200">
                {warnings.length == 1
                  ? `${warnings.length} Warning`
                  : `${warnings.length} Warnings`}
              </Badge>
            )}
          </div>
          <div className="flex">
            {onValidate && (
              <Button
                variant="ghost"
                className="h-4 w-4"
                size="icon"
                key="validate"
                type="button"
              >
                {isValidating ? (
                  <Spinner className="h-4 w-4" />
                ) : (
                  <ReloadIcon
                    className="h-4 w-4"
                    onClick={() => onValidate()}
                  />
                )}
              </Button>
            )}
          </div>
        </div>
        <CardDescription>
          A list of schema validation errors to resolve before moving forward.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col flex-1">
        {formErrors.length === 0 && warnings.length === 0 ? (
          <div className="flex flex-col flex-1 items-center justify-center bg-green-100 dark:bg-green-900 text-green-900 dark:text-green-200 rounded-xl">
            <div className="text-sm flex flex-row items-center gap-2 px-1">
              <div className="flex">
                <CheckIcon />
              </div>
              <p>Everything looks good!</p>
            </div>
          </div>
        ) : (
          <ScrollArea className="max-h-[177px] overflow-auto">
            <div className="flex flex-col gap-2">
              {messages.map((message, index) => (
                <div
                  key={message + index}
                  className="text-xs bg-red-200 dark:bg-red-800/70 rounded-sm p-2 text-wrap"
                >
                  {message}
                </div>
              ))}
              {warnings.map((message, index) => (
                <div
                  key={message + index}
                  className="text-xs bg-yellow-200 dark:bg-yellow-800/70 rounded-sm p-2 text-wrap"
                >
                  {message}
                </div>
              ))}
            </div>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  );
}

function formErrorsToMessages(errors: FormError[]): string[] {
  const messages: string[] = [];
  errors.forEach((error) => {
    if (error.level !== 'error') {
      return;
    }
    const pieces: string[] = [error.path];
    if (error.type) {
      pieces.push(`[${error.type}]`);
    }
    pieces.push(error.message);
    messages.push(pieces.join(' '));
  });

  return messages;
}

function formWarningsToMessages(warnings: FormError[]): string[] {
  const messages: string[] = [];
  warnings.forEach((warning) => {
    if (warning.level !== 'warning') {
      return;
    }
    const pieces: string[] = [warning.path];
    if (warning.type) {
      pieces.push(`[${warning.type}]`);
    }
    pieces.push(warning.message);
    messages.push(pieces.join(' '));
  });

  return messages;
}
