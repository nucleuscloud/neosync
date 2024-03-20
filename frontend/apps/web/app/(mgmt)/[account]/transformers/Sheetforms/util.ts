export interface TransformerFormProps<T> {
  existingConfig?: T;
  onSubmit(config: T): void;
  isReadonly?: boolean;
}

export function setBigIntOrOld(
  newVal: bigint | boolean | number | string,
  oldValue: bigint
): bigint {
  try {
    const newInt = BigInt(newVal);
    return newInt;
  } catch {
    return oldValue;
  }
}
