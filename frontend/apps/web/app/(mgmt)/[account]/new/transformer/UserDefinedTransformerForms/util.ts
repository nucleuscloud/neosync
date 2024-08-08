import { FieldErrors, FieldValues } from 'react-hook-form';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export interface TransformerConfigProps<
  T extends Record<string, any>,
  TError extends FieldValues = T,
> {
  value: T;
  setValue(newValue: T): void;
  isDisabled?: boolean;

  errors?: FieldErrors<TError>;
}
