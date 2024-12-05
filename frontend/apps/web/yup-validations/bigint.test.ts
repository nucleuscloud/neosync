/**
 * @jest-environment node
 */

import {
  getBigIntValidateMaxFn,
  getBigIntValidateMinFn,
  getBigIntValidator,
} from './bigint';

describe('getBigIntValidateMaxFn', () => {
  it('should return true for values less than or equal to the maximum', () => {
    const numberValidator = getBigIntValidateMaxFn(10);
    const stringValidator = getBigIntValidateMaxFn('10');
    const bigIntValidator = getBigIntValidateMaxFn(BigInt(10));

    expect(numberValidator(BigInt(5))).toBe(true);
    expect(numberValidator(undefined)).toBe(true);

    expect(stringValidator(BigInt(5))).toBe(true);
    expect(stringValidator(undefined)).toBe(true);

    expect(bigIntValidator(BigInt(5))).toBe(true);
    expect(bigIntValidator(undefined)).toBe(true);
  });

  it('should return false for values greater than the maximum', () => {
    const numberValidator = getBigIntValidateMaxFn(10);
    const stringValidator = getBigIntValidateMaxFn('10');
    const bigIntValidator = getBigIntValidateMaxFn(BigInt(10));

    expect(numberValidator(BigInt(20))).toBe(false);
    expect(numberValidator(undefined)).toBe(true);

    expect(stringValidator(BigInt(20))).toBe(false);
    expect(stringValidator(undefined)).toBe(true);

    expect(bigIntValidator(BigInt(20))).toBe(false);
    expect(bigIntValidator(undefined)).toBe(true);
  });

  it('should handle default values', () => {
    const maxValidatorWithDefault = getBigIntValidateMaxFn(undefined);

    expect(maxValidatorWithDefault(BigInt(5))).toBe(false);
    expect(maxValidatorWithDefault(undefined)).toBe(true);
  });

  it('should return false for undefined maxval values', () => {
    const undefinedValidator = getBigIntValidateMaxFn(undefined);

    expect(undefinedValidator(BigInt(5))).toBe(false);
    expect(undefinedValidator(undefined)).toBe(true);
  });
});

describe('getBigIntValidateMinFn', () => {
  it('should return false for values less than or equal to the maximum', () => {
    const numberValidator = getBigIntValidateMinFn(10);
    const stringValidator = getBigIntValidateMinFn('10');
    const bigIntValidator = getBigIntValidateMinFn(BigInt(10));

    expect(numberValidator(BigInt(5))).toBe(false);
    expect(numberValidator(undefined)).toBe(false);

    expect(stringValidator(BigInt(5))).toBe(false);
    expect(stringValidator(undefined)).toBe(false);

    expect(bigIntValidator(BigInt(5))).toBe(false);
    expect(bigIntValidator(undefined)).toBe(false);
  });

  it('should return true for values greater than the maximum', () => {
    const numberValidator = getBigIntValidateMinFn(10);
    const stringValidator = getBigIntValidateMinFn('10');
    const bigIntValidator = getBigIntValidateMinFn(BigInt(10));

    expect(numberValidator(BigInt(20))).toBe(true);
    expect(numberValidator(undefined)).toBe(false);

    expect(stringValidator(BigInt(20))).toBe(true);
    expect(stringValidator(undefined)).toBe(false);

    expect(bigIntValidator(BigInt(20))).toBe(true);
    expect(bigIntValidator(undefined)).toBe(false);
  });

  it('should handle default values', () => {
    const maxValidatorWithDefault = getBigIntValidateMinFn(undefined);

    expect(maxValidatorWithDefault(BigInt(5))).toBe(true);
    expect(maxValidatorWithDefault(undefined)).toBe(true);
  });

  it('should return true for undefined maxval values', () => {
    const undefinedValidator = getBigIntValidateMinFn(undefined);

    expect(undefinedValidator(BigInt(5))).toBe(true);
    expect(undefinedValidator(undefined)).toBe(true);
  });
});

describe('getBigIntValidator', () => {
  function createValidator(
    range: [number | string | bigint, number | string | bigint],
    defaultValue?: number | string | bigint,
    requiredMessage?: string
  ) {
    return getBigIntValidator({
      range,
      default: defaultValue,
      requiredMessage,
    });
  }

  it('should validate a value within the range', async () => {
    const validator = createValidator([BigInt(10), BigInt(100)]);

    await expect(validator.validate(BigInt(50))).resolves.toBe(BigInt(50));
    await expect(validator.validate(BigInt(10))).resolves.toBe(BigInt(10));
    await expect(validator.validate(BigInt(100))).resolves.toBe(BigInt(100));
  });

  it('should invalidate a value outside the range', async () => {
    const validator = createValidator([BigInt(10), BigInt(100)]);

    await expect(validator.validate(BigInt(5))).rejects.toThrow(
      'Value must be greater than or equal to 10'
    );
    await expect(validator.validate(BigInt(101))).rejects.toThrow(
      'Value must be less than or equal to 100'
    );
  });

  it('should use default value if input is defined', async () => {
    const defaultValue = BigInt(50);
    const validator = createValidator([BigInt(10), BigInt(100)], defaultValue);

    await expect(validator.validate(BigInt(50))).resolves.toBe(defaultValue);
  });

  it('should return required error message if value is not provided and requiredMessage is set', async () => {
    const requiredMessage = 'Value must be greater than or equal to 10';
    const validator = createValidator(
      [BigInt(10), BigInt(100)],
      undefined,
      requiredMessage
    );

    await expect(validator.validate(undefined)).rejects.toThrow(
      requiredMessage
    );
  });
});
