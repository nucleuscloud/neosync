import TruncatedText from '@/components/TruncatedText';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/libs/utils';
import { CheckIcon, Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { NosqlJobMappingRow } from '../JobMappingTable/Columns';

interface Props {
  text: string;
  onEdit(updatedObject: Pick<NosqlJobMappingRow, 'column'>): void;
  isDuplicate(val: string, currValue?: string): boolean;
}

export default function EditDocumentKey(props: Props): ReactElement<any> {
  const { text, onEdit, isDuplicate } = props;
  const [isEditingMapping, setIsEditingMapping] = useState(false);
  const [inputValue, setInputValue] = useState<string>(text);
  const [duplicateError, setDuplicateError] = useState(false);

  const handleSave = () => {
    onEdit({ column: inputValue });
    setIsEditingMapping(false);
  };

  const handleDocumentKeyChange = (val: string) => {
    setInputValue(val);
    setDuplicateError(isDuplicate(val, text));
  };

  return (
    <div className="w-full flex flex-row items-center gap-1">
      {isEditingMapping ? (
        <>
          <Input
            value={inputValue}
            onChange={(e) => handleDocumentKeyChange(e.target.value)}
            className={cn(duplicateError ? 'border border-red-400 ring-' : '')}
          />
          <div className="text-red-400 text-xs pl-2">
            {duplicateError && 'Already exists'}
          </div>
        </>
      ) : (
        <TruncatedText text={inputValue} />
      )}
      <Button
        variant="outline"
        size="sm"
        className="hidden h-[36px] lg:flex"
        type="button"
        disabled={duplicateError}
        onClick={() => {
          if (isEditingMapping) {
            handleSave();
          } else {
            setIsEditingMapping(true);
          }
        }}
      >
        {isEditingMapping ? <CheckIcon /> : <Pencil1Icon />}
      </Button>
    </div>
  );
}
