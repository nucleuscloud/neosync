import { FieldErrors, FieldValues } from 'react-hook-form';

export interface TransformerConfigProps<
  T extends Record<string, any>, // eslint-disable-line @typescript-eslint/no-explicit-any
  TError extends FieldValues = T,
> {
  value: T;
  setValue(newValue: T): void;
  isDisabled?: boolean;

  errors?: FieldErrors<TError>;
}
