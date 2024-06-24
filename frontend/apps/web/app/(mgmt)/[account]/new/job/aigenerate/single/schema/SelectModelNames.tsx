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
  OpenAI: ['gpt-3.5-turbo', 'gpt-4', 'gpt-4-32k', 'gpt-4-turbo', 'gpt-4o'],
};

export default function SelectModelNames(props: Props): ReactElement {
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
