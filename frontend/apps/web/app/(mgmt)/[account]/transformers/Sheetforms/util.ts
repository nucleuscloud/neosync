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

export function tryBigInt(
  val: bigint | boolean | number | string
): bigint | null {
  try {
    const newInt = BigInt(val);
    return newInt;
  } catch {
    return null;
  }
}
