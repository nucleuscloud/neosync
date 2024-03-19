export interface TransformerFormProps<T> {
  existingConfig?: T;
  onSubmit(config: T): void;
  isReadonly?: boolean;
}
