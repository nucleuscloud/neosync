import { TestContext, ValidationError } from 'yup';
import { parseDuration } from '../util/duration';
// Ensures that the number being validated is always greater than or equal to a minimum value
export function getNumberValidateMinFn(
  minVal: number
): (value: number | undefined) => boolean {
  const convertedMinValue = numberOrDefault(minVal, 0);
  return (value) => {
    const maxValue = numberOrDefault(value, 0);
    return maxValue >= convertedMinValue;
  };
}

// Ensures that the number being validated is always less than or equal to a maximum value
export function getNumberValidateMaxFn(
  maxVal: number
): (value: number | undefined) => boolean {
  const convertedMaxValue = numberOrDefault(maxVal, 0);
  return (value) => {
    const minValue = numberOrDefault(value, 0);
    return minValue <= convertedMaxValue;
  };
}

function numberOrDefault(
  value: number | null | undefined,
  defaultVal: number
): number {
  return value == null ? defaultVal : value;
}

// Allows empty duration values
export function getDurationValidateFn(): (
  value: string | undefined,
  context: TestContext<unknown>
) => boolean | ValidationError {
  return (value, context) => {
    if (!value) {
      return true;
    }
    try {
      parseDuration(value);
      return true;
    } catch (err) {
      return context.createError({
        message:
          err instanceof Error ? err.message : 'Must be a valid duration',
      });
    }
  };
}
