import { ReactElement } from 'react';
import { Input, InputProps } from './ui/input';

interface NumberedInputProps extends Omit<InputProps, 'onChange'> {
  onChange(value: number): void;
}

export default function NumberedInput(props: NumberedInputProps): ReactElement {
  const { onChange, ...rest } = props;

  return (
    <Input
      {...rest}
      type="number"
      onChange={(event) => {
        const numVal = event.target.valueAsNumber;
        if (!isNaN(numVal)) onChange(numVal);
      }}
    />
  );
}
