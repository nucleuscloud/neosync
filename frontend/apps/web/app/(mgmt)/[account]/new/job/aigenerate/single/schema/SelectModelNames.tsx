import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ReactElement, useState } from 'react';

interface Props {
  onSelected(value: string): void;
}

// Add more common models here to make it easier for folks to plug in to well known OpenAI compatible platforms
const COMMON_MODELS = {
  OpenAI: ['gpt-4o-mini', 'gpt-4o', 'gpt-4-turbo'],
};

export default function SelectModelNames(props: Props): ReactElement<any> {
  const { onSelected } = props;
  const [value, setValue] = useState('');
  return (
    <Select
      onValueChange={(value) => {
        setValue(value);
        onSelected(value);
      }}
      defaultValue={value}
    >
      <SelectTrigger>
        <SelectValue placeholder="Common model names..." />
      </SelectTrigger>
      <SelectContent>
        {Object.entries(COMMON_MODELS).map(([group, modelNames]) => (
          <SelectGroup key={group}>
            <SelectLabel>{group}</SelectLabel>
            {modelNames.map((modelName) => (
              <SelectItem key={`${group}-${modelName}`} value={modelName}>
                {modelName}
              </SelectItem>
            ))}
          </SelectGroup>
        ))}
      </SelectContent>
    </Select>
  );
}
