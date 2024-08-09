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
