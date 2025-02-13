import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { Input } from '@/components/ui/input';
import { ReactElement } from 'react';

interface NameProps {
  error?: string;
  value: string;
  onChange(value: string): void;
}

export function Name(props: NameProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="name"
        title="Name"
        description="Name of the connection for display and reference, must be unique"
        isErrored={!!error}
      />
      <Input
        id="name"
        autoCapitalize="off" // we don't allow capitals
        data-1p-ignore // tells 1password extension to not autofill this field
        value={value || ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Connection name"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}
