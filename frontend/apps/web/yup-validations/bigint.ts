import * as Yup from 'yup';

type BigIntVal = number | string | bigint;

interface BigIntValidatorConfig {
  default?: BigIntVal;
  requiredMessage?: string;
  // Inclusive range [min, max]
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
      'is-bigint-like',
      'Value must be bigint or convertable to bigint',
      isBigIntLike
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

// Checks the input value to see if it could be converted to a bigint.
// If the input is nullish, it's still considered so as that is used as a default fallback of 0.
function isBigIntLike(
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

// Ensures that the bigint being validated is always less than or equal to a minimum value
export function getBigIntValidateMinFn(
  minVal: number | string | bigint | undefined
): (value: bigint | undefined) => boolean {
  const convertedMinValue = bigIntOrDefault(minVal, BigInt(0));
  return (value) => {
    const maxVal = bigIntOrDefault(value, BigInt(0));
    return maxVal >= convertedMinValue;
  };
}
// Ensures that the bigint being validated is always less than or equal to a maximum value
export function getBigIntValidateMaxFn(
  maxVal: number | string | bigint | undefined
): (value: bigint | undefined) => boolean {
  const convertedMaxValue = bigIntOrDefault(maxVal, BigInt(0));
  return (value) => {
    const minVal = bigIntOrDefault(value, BigInt(0));
    return minVal <= convertedMaxValue;
  };
}

// Returns the input as a bigint, or falls back to the default if it isn't convertable
function bigIntOrDefault(value: unknown, defaultVal: bigint): bigint {
  if (isBigIntLike(value) && value != null) {
    return BigInt(value);
  }
  return defaultVal;
}
