/**
 * @jest-environment node
 */

import { getNumberValidateMaxFn, getNumberValidateMinFn } from './number';

describe('getNumberValidateMaxFn', () => {
  it('should return true for values less than the maximum', () => {
    const validator = getNumberValidateMaxFn(10);

    expect(validator(5)).toBe(true);
  });

  it('should return false for values greater than or equal to the maximum', () => {
    const validator = getNumberValidateMaxFn(10);

    expect(validator(10)).toBe(true);
    expect(validator(20)).toBe(false);
  });

  it('should handle default minimum value as 0', () => {
    const validator = getNumberValidateMaxFn(undefined as unknown as number);

    expect(validator(5)).toBe(false);
    expect(validator(-1)).toBe(true);
  });

  it('should return true for undefined values', () => {
    const validator = getNumberValidateMaxFn(10);

    expect(validator(undefined)).toBe(true);
  });
});

describe('getNumberValidateMinFn', () => {
  it('should return false for values less than the minimum', () => {
    const validator = getNumberValidateMinFn(10);

    expect(validator(5)).toBe(false);
  });

  it('should return true for values greater than or equal to the minimum', () => {
    const validator = getNumberValidateMinFn(10);

    expect(validator(10)).toBe(true);
    expect(validator(20)).toBe(true);
  });

  it('should handle default minimum value as 0', () => {
    const validator = getNumberValidateMinFn(undefined as unknown as number);

    expect(validator(5)).toBe(true);
    expect(validator(-1)).toBe(false);
  });

  it('should return true for undefined values', () => {
    const validator = getNumberValidateMinFn(10);

    expect(validator(undefined)).toBe(false);
  });
});
