import { format } from 'date-fns';

export function formatDateTime(
  dateStr?: string | Date | number,
  is24Hour = false
): string | undefined {
  if (!dateStr) {
    return undefined;
  }
  const hourFormat = is24Hour ? 'HH' : 'hh';
  const amPm = is24Hour ? '' : 'a';
  return format(new Date(dateStr), `MM/dd/yyyy ${hourFormat}:mm:ss ${amPm}`);
}

export function formatDateTimeMilliseconds(
  dateStr?: string | Date | number,
  is24Hour = false
): string | undefined {
  if (!dateStr) {
    return undefined;
  }
  const hourFormat = is24Hour ? 'HH' : 'hh';
  const amPm = is24Hour ? '' : 'a';

  return format(
    new Date(dateStr),
    `MM/dd/yyyy ${hourFormat}:mm:ss:SSS ${amPm}`
  );
}

export function getErrorMessage(error: unknown): string {
  return isErrorWithMessage(error) ? error.message : 'unknown error message';
}
function isErrorWithMessage(error: unknown): error is { message: string } {
  return (
    typeof error === 'object' &&
    error != null &&
    'message' in error &&
    typeof error.message === 'string'
  );
}

export const toTitleCase = (str: string) => {
  return str.toLowerCase().replace(/\b\w/g, (s) => s.toUpperCase());
};

const NANOS_PER_SECOND = BigInt(1000000000);
const SECONDS_PER_MIN = BigInt(60);

// if the duration is too large to convert to minutes, it will return the max safe integer
export function convertNanosecondsToMinutes(duration: bigint): number {
  // Convert nanoseconds to minutes
  const minutesBigInt = duration / NANOS_PER_SECOND / SECONDS_PER_MIN;

  // Check if the result is within the safe range for JavaScript numbers
  if (minutesBigInt <= BigInt(Number.MAX_SAFE_INTEGER)) {
    return Number(minutesBigInt);
  } else {
    // Handle the case where the number of minutes is too large
    console.warn(
      'The number of minutes is too large for a safe JavaScript number. Returning as BigInt.'
    );
    return Number.MAX_SAFE_INTEGER;
  }
}

// Convert minutes to BigInt to ensure precision in multiplication
export function convertMinutesToNanoseconds(minutes: number): bigint {
  const minutesBigInt = BigInt(minutes);
  return minutesBigInt * SECONDS_PER_MIN * NANOS_PER_SECOND;
}
