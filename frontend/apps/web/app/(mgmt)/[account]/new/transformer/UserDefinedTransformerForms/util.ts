export interface TransformerConfigProps<T> {
  value: T;
  setValue(newValue: T): void;
  isDisabled?: boolean;
}
