import TruncatedText from '@/components/TruncatedText';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { CheckIcon, Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

interface Props {
  collections: string[];
  text: string;
  onEdit(updatedObject: { collection: string }): void;
}

export default function EditCollection(props: Props): ReactElement {
  const { text, collections, onEdit } = props;

  const [isEditingMapping, setIsEditingMapping] = useState(false);
  const [isSelectedCollection, setSelectedCollection] = useState<string>(text);

  const handleSave = () => {
    onEdit({ collection: isSelectedCollection });
    setIsEditingMapping(false);
  };

  return (
    <div className="w-full flex flex-row items-center gap-1">
      {isEditingMapping ? (
        <Select
          onValueChange={(val) => setSelectedCollection(val)}
          value={isSelectedCollection}
        >
          <SelectTrigger>
            <SelectValue
              placeholder="Select a collection"
              className="placeholder:text-muted-foreground/70"
            />
          </SelectTrigger>
          <SelectContent>
            {collections.map((collection) => (
              <SelectItem value={collection} key={collection}>
                {collection}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ) : (
        <TruncatedText text={isSelectedCollection} />
      )}
      <Button
        variant="outline"
        size="sm"
        className="hidden h-[36px] lg:flex"
        type="button"
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
