import { format } from 'date-fns';

interface Instance {
  createdDate?: string;
}

export function formatDate(dateStr?: string): string | undefined {
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

export function formatInstanceCreatedDate<T extends Instance>(
  instances?: T[]
): T[] {
  if (!instances) {
    return [];
  }
  return instances.map((instance) => {
    return {
      ...instance,
      createdDate: instance.createdDate
        ? formatDate(instance.createdDate)
        : undefined,
    };
  });
}

// replaces / and spaces with -
export function formatUrlParam(str?: string): string {
  if (!str) {
    return '';
  }
  const regex = /[\/ ]/gi;
  return str.replaceAll(regex, '-').toLocaleLowerCase();
}

export function getClientTz(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

export function containsWhiteSpace(str: string): boolean {
  return /\s/g.test(str);
}

export function titleCase(str: string): string {
  return str[0].toUpperCase() + str.slice(1).toLowerCase();
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

export function getSingleOrUndefined(
  item: string | string[] | undefined
): string | undefined {
  if (!item) {
    return undefined;
  }
  const newItem = Array.isArray(item) ? item[0] : item;
  return !newItem || newItem === 'undefined' ? undefined : newItem;
}
