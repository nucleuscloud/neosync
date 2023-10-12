import { format } from 'date-fns';

interface Instance {
  createdDate?: string;
}

function formatDate(dateStr?: string): string | undefined {
  if (!dateStr) {
    return undefined;
  }
  return format(new Date(dateStr), 'MM/dd/yyyy');
}

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
    error !== null &&
    'message' in error &&
    typeof error.message === 'string'
  );
}
