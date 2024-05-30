import {
  GenerateEmailType,
  InvalidEmailAction,
  SupportedJobType,
  TransformerDataType,
  TransformerSource,
} from '@neosync/sdk';
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

function getTransformerDataTypeString(dt: TransformerDataType): string {
  const value = TransformerDataType[dt];
  return value ? value.toLowerCase() : 'unspecified';
}

export function getTransformerDataTypesString(
  dts: TransformerDataType[]
): string {
  return dts.map((dt) => getTransformerDataTypeString(dt)).join(' | ');
}

function getTransformerJobTypeString(dt: SupportedJobType): string {
  const value = SupportedJobType[dt];
  return value ? toTitleCase(value) : 'unspecified';
}

export function getTransformerJobTypesString(
  dts: SupportedJobType[]
): string[] {
  return dts.map((dt) => getTransformerJobTypeString(dt));
}

export function getTransformerSourceString(ds: TransformerSource): string {
  const value = TransformerSource[ds];
  return value ? value.toLowerCase() : 'unspecified';
}

export function getGenerateEmailTypeString(
  emailType: GenerateEmailType
): string {
  const value = GenerateEmailType[emailType];
  return value ? value.toLowerCase() : 'unknown';
}

// This expects the fully qualified proto email that looks like this: GENERATE_EMAIL_TYPE_FULLNAME
// This is because the form deals with the TransformConfig that is jsonified and contains the wired value instead of the int value.
// There seems to be no easy way to convert between the two
export function generateEmailTypeStringToEnum(
  emailType: string
): GenerateEmailType {
  switch (emailType) {
    case 'GENERATE_EMAIL_TYPE_FULLNAME':
      return GenerateEmailType.FULLNAME;
    default:
      return GenerateEmailType.UUID_V4;
  }
}

export function getInvalidEmailActionString(
  invalidEmailAction: InvalidEmailAction
): string {
  const value = InvalidEmailAction[invalidEmailAction];
  return value ? value.toLowerCase() : 'unknown';
}

// This expects the fully qualified proto email that looks like this: INVALID_EMAIL_ACTION_REJECT
// This is because the form deals with the TransformConfig that is jsonified and contains the wired value instead of the int value.
// There seems to be no easy way to convert between the two
export function invalidEmailActionStringToEnum(
  invalidEmailAction: string
): InvalidEmailAction {
  switch (invalidEmailAction) {
    case 'INVALID_EMAIL_ACTION_PASSTHROUGH':
      return InvalidEmailAction.PASSTHROUGH;
    case 'INVALID_EMAIL_ACTION_GENERATE':
      return InvalidEmailAction.GENERATE;
    case 'INVALID_EMAIL_ACTION_NULL':
      return InvalidEmailAction.NULL;
    default:
      return InvalidEmailAction.REJECT;
  }
}
