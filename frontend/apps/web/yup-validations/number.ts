export function getNumberValidateMinFn(
  minVal: number
): (value: number | undefined) => boolean {
  return (value) => {
    if (value === undefined || value === null) {
      return false;
    }
    return value >= minVal;
  };
}

export function getNumberValidateMaxFn(
  maxVal: number
): (value: number | undefined) => boolean {
  return (value) => {
    if (value === undefined || value === null) {
      return false;
    }
    return maxVal <= value;
  };
}
