export function getNumberValidateMinFn(
  minVal: number
): (value: number | undefined) => boolean {
  return (value) => {
    const maxValue = numberOrDefault(value, 0);
    const convertedMinValue = numberOrDefault(minVal, 0);
    return maxValue >= convertedMinValue;
  };
}

export function getNumberValidateMaxFn(
  maxVal: number
): (value: number | undefined) => boolean {
  return (value) => {
    const minValue = numberOrDefault(value, 0);
    const convertedMaxValue = numberOrDefault(maxVal, 0);
    return minValue <= convertedMaxValue;
  };
}

function numberOrDefault(
  value: number | null | undefined,
  defaultVal: number
): number {
  return value == null ? defaultVal : value;
}
