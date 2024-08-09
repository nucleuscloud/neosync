import * as Yup from 'yup';

type BigIntVal = number | string | bigint;
export interface BigIntValidatorConfig {
  default?: BigIntVal;
  requiredMessage?: string;
  range: [BigIntVal, BigIntVal];
}
export function getBigIntValidator(
  config: BigIntValidatorConfig
): Yup.MixedSchema<
  bigint | undefined,
  Yup.AnyObject,
  bigint | undefined,
  'd' | ''
> {
  const minValidator = getBigIntValidateMinFn(config.range[0]);
  const maxValidator = getBigIntValidateMaxFn(config.range[1]);
  const validator = Yup.mixed<bigint>()
    .test(
      'is-bigint',
      'Value must be bigint or convertable to bigint',
      isBigIntable
    )
    .test(
      'min-value',
      `Value must be greater than or equal to ${config.range[0]}`,
      minValidator
    )
    .test(
      'max-value',
      `Value must be less than or equal to ${config.range[1]}`,
      maxValidator
    );
  if (config.default != null) {
    validator.default(BigInt(config.default));
  }
  if (config.requiredMessage != null) {
    validator.required(config.requiredMessage);
  }
  return validator;
}

function isBigIntable(
  value: unknown
): value is bigint | number | string | null | undefined {
  if (typeof value === 'bigint') {
    return true;
  } else if (typeof value === 'number') {
    return true;
  } else if (typeof value === 'string') {
    try {
      BigInt(value);
      return true;
    } catch {
      return false;
    }
  } else if (value == null) {
    return true;
  }
  return false;
}

// BigInt validator for Minimum values
export function getBigIntValidateMinFn(
  minVal: number | string | bigint | undefined
): (value: bigint | undefined) => boolean {
  return (value) => {
    const maxVal = bigIntOrDefault(value, BigInt(0));
    const convertedMinValue = bigIntOrDefault(minVal, BigInt(0));
    return maxVal >= convertedMinValue;
  };
}
// BigInt validator for Maximum values
export function getBigIntValidateMaxFn(
  maxVal: number | string | bigint | undefined
): (value: bigint | undefined) => boolean {
  return (value) => {
    const minVal = bigIntOrDefault(value, BigInt(0));
    const convertedMaxValue = bigIntOrDefault(maxVal, BigInt(0));
    return minVal <= convertedMaxValue;
  };
}

function bigIntOrDefault(value: unknown, defaultVal: bigint): bigint {
  if (isBigIntable(value) && value != null) {
    return BigInt(value);
  }
  return defaultVal;
}
