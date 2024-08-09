import * as Yup from 'yup';

type BigIntVal = number | string | bigint;
export interface BigIntValidatorConfig {
  default: BigIntVal;
  requiredMessage: string;
  range: [BigIntVal, BigIntVal];
}
export function getBigIntValidator(
  config: BigIntValidatorConfig
): Yup.MixedSchema<bigint, Yup.AnyObject, bigint, 'd'> {
  const minValidator = getBigIntValidateMinFn(config.range[0]);
  const maxValidator = getBigIntValidateMaxFn(config.range[1]);
  return Yup.mixed<bigint>()
    .default(BigInt(config.default))
    .required(config.requiredMessage)
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
}

function isBigIntable(value: unknown): value is bigint | number | string {
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
  }
  return false;
}

// BigInt validator for Minimum values
export function getBigIntValidateMinFn(
  minVal: number | string | bigint
): (value: bigint | undefined) => boolean {
  return (value) => {
    if (value === undefined || value === null) {
      return false;
    }
    const convertedMinValue = BigInt(minVal);
    try {
      const bigIntValue = BigInt(value);
      return bigIntValue >= convertedMinValue;
    } catch {
      return false; // Not convertible to BigInt, but this should theoretically not happen due to previous test
    }
  };
}
// BigInt validator for Maximum values
export function getBigIntValidateMaxFn(
  maxVal: number | string | bigint
): (value: bigint | undefined) => boolean {
  return (value) => {
    if (value === undefined || value === null) {
      return false;
    }
    const convertedMaxValue = BigInt(maxVal);
    try {
      const bigIntValue = BigInt(value);
      return bigIntValue <= convertedMaxValue;
    } catch {
      return false; // Not convertible to BigInt, but this should theoretically not happen due to previous test
    }
  };
}
