/**
 * @jest-environment node
 */

import {
  GenerateEmailType,
  InvalidEmailAction,
  SupportedJobType,
  TransformerDataType,
  TransformerSource,
} from '@neosync/sdk';
import { format } from 'date-fns';
import {
  convertMinutesToNanoseconds,
  convertNanosecondsToMinutes,
  formatDateTime,
  formatDateTimeMilliseconds,
  getErrorMessage,
  getGenerateEmailTypeString,
  getInvalidEmailActionString,
  getTransformerDataTypesString,
  getTransformerJobTypesString,
  getTransformerSourceString,
  toTitleCase,
} from './util';

describe('formatDateTime', () => {
  it('should format a date string in 12-hour format', () => {
    const dateStr = '2024-08-16T14:30:00Z';
    const result = formatDateTime(dateStr);
    const expected = format(new Date(dateStr), 'MM/dd/yyyy hh:mm:ss a');
    expect(result).toBe(expected);
  });

  it('should format a date string in 24-hour format', () => {
    const dateStr = '2024-08-16T14:30:00Z';
    const result = formatDateTime(dateStr, true);
    const expected = format(new Date(dateStr), 'MM/dd/yyyy HH:mm:ss');
    expect(result).toBe(expected);
  });

  it('should handle a Date object input', () => {
    const dateObj = new Date('2024-08-16T14:30:00Z');
    const result = formatDateTime(dateObj);
    const expected = format(dateObj, 'MM/dd/yyyy hh:mm:ss a');
    expect(result).toBe(expected);
  });

  it('should handle a timestamp input', () => {
    const timestamp = new Date('2024-08-16T14:30:00Z').getTime();
    const result = formatDateTime(timestamp);
    const expected = format(new Date(timestamp), 'MM/dd/yyyy hh:mm:ss a');
    expect(result).toBe(expected);
  });

  it('should return undefined for no input', () => {
    const result = formatDateTime();
    expect(result).toBeUndefined();
  });

  it('should return undefined for invalid date input', () => {
    const result = formatDateTime('invalid-date');
    expect(result).toBeUndefined();
  });
});

describe('formatDateTimeMilliseconds', () => {
  it('should format a date string in 12-hour format', () => {
    const dateStr = '2024-08-16T14:30:00Z';
    const result = formatDateTimeMilliseconds(dateStr);
    const expected = format(new Date(dateStr), 'MM/dd/yyyy hh:mm:ss:SSS a');
    expect(result).toBe(expected);
  });

  it('should format a date string in 24-hour format', () => {
    const dateStr = '2024-08-16T14:30:00Z';
    const result = formatDateTimeMilliseconds(dateStr, true);
    const expected = format(new Date(dateStr), 'MM/dd/yyyy HH:mm:ss:SSS');
    expect(result).toBe(expected);
  });

  it('should handle a Date object input', () => {
    const dateObj = new Date('2024-08-16T14:30:00Z');
    const result = formatDateTimeMilliseconds(dateObj);
    const expected = format(dateObj, 'MM/dd/yyyy hh:mm:ss:SSS a');
    expect(result).toBe(expected);
  });

  it('should handle a timestamp input', () => {
    const timestamp = new Date('2024-08-16T14:30:00Z').getTime();
    const result = formatDateTimeMilliseconds(timestamp);
    const expected = format(new Date(timestamp), 'MM/dd/yyyy hh:mm:ss:SSS a');
    expect(result).toBe(expected);
  });

  it('should return undefined for no input', () => {
    const result = formatDateTimeMilliseconds();
    expect(result).toBeUndefined();
  });

  it('should return undefined for invalid date input', () => {
    const result = formatDateTimeMilliseconds('invalid-date');
    expect(result).toBeUndefined();
  });
});

describe('getErrorMessage', () => {
  it('should return an error message when the error has a message property', () => {
    const error = { message: 'error message' };
    const result = getErrorMessage(error);
    expect(result).toBe('error message');
  });

  it('should return an unknown error message when the error message does not have a message property', () => {
    const error = { status: 500 };
    const result = getErrorMessage(error);
    expect(result).toBe('unknown error message');
  });

  it('should return an unknown error message when the error is a string', () => {
    const error = 'error';
    const result = getErrorMessage(error);
    expect(result).toBe('unknown error message');
  });

  it('should return an unknown error message when the error is a number', () => {
    const error = 500;
    const result = getErrorMessage(error);
    expect(result).toBe('unknown error message');
  });

  it('should return an unknown error message when the error is null', () => {
    const error = null;
    const result = getErrorMessage(error);
    expect(result).toBe('unknown error message');
  });

  it('should return an unknown error message when the error is undefined', () => {
    const error = undefined;
    const result = getErrorMessage(error);
    expect(result).toBe('unknown error message');
  });
});

describe('toTitleCase', () => {
  it('should return lowercase string in title case', () => {
    const input = 'hello world';
    const result = toTitleCase(input);
    expect(result).toBe('Hello World');
  });

  it('should return uppercase string in title case', () => {
    const input = 'HEllO WORLD';
    const result = toTitleCase(input);
    expect(result).toBe('Hello World');
  });
  it('should return titlecase string in title case', () => {
    const input = 'Hello World';
    const result = toTitleCase(input);
    expect(result).toBe('Hello World');
  });
});

describe('convertNanosecondsToMinutes', () => {
  it('should convert nanoseconds to 1 min', () => {
    const nano = BigInt(60000000000);
    const result = convertNanosecondsToMinutes(nano);
    expect(result).toEqual(1);
  });
  it('should convert  nanoseconds to 0 min', () => {
    const nano = BigInt(6000000);
    const result = convertNanosecondsToMinutes(nano);
    expect(result).toEqual(0);
  });
  it('should return max safe integer', () => {
    const nano = BigInt(6000000000000000000000000000000);
    const result = convertNanosecondsToMinutes(nano);
    expect(result).toEqual(Number.MAX_SAFE_INTEGER);
  });
});

describe('convertMinutesToNanoseconds', () => {
  it('should convert 2 min to nanoseconds', () => {
    const min = 2;
    const result = convertMinutesToNanoseconds(min);
    expect(result).toEqual(BigInt(2) * BigInt(1000000000) * BigInt(60));
  });
  it('should convert 0 min to 0 nanoseconds', () => {
    const min = 0;
    const result = convertMinutesToNanoseconds(min);
    expect(result).toEqual(BigInt(0) * BigInt(1000000000) * BigInt(60));
  });
});

describe('getTransformerDataTypesString', () => {
  it('should return the correct string for a single element array', () => {
    const value = [TransformerDataType.STRING];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('string');
  });
  it('should return string for string data types', () => {
    const value = [TransformerDataType.STRING, TransformerDataType.STRING];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('string | string');
  });
  it('should return string for string and int64 data types', () => {
    const value = [TransformerDataType.STRING, TransformerDataType.INT64];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('string | int64');
  });
  it('should return string for boolean and float data types', () => {
    const value = [TransformerDataType.BOOLEAN, TransformerDataType.FLOAT64];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('boolean | float64');
  });
  it('should return string for null and time data types', () => {
    const value = [TransformerDataType.NULL, TransformerDataType.TIME];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('null | time');
  });
  it('should return string for any and uuid data types', () => {
    const value = [
      TransformerDataType.ANY,
      TransformerDataType.UUID,
      TransformerDataType.UNSPECIFIED,
    ];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('any | uuid | unspecified');
  });
  it('should return empty string if empty array', () => {
    const value: TransformerDataType[] = [];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('');
  });
  it('should return "unspecified" for any invalid values while returning the right value for valid values', () => {
    const invalid = 'invalid' as unknown as TransformerDataType;
    const value = [TransformerDataType.STRING, invalid];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('string | unspecified');
  });
  it('should return "unspecified" for an array with invalid types', () => {
    const value = [
      100,
      'invalid',
      undefined,
    ] as unknown as TransformerDataType[];
    const result = getTransformerDataTypesString(value);
    expect(result).toBe('unspecified | unspecified | unspecified');
  });
});

describe('getTransformerJobTypesString', () => {
  it('should return the correct titlecase string array for a single element array', () => {
    const value = [SupportedJobType.GENERATE];
    const result = getTransformerJobTypesString(value);
    expect(result).toStrictEqual(['Generate']);
  });
  it('should return a titlecase string array for string data types', () => {
    const value = [SupportedJobType.SYNC, SupportedJobType.GENERATE];
    const result = getTransformerJobTypesString(value);
    expect(result).toStrictEqual(['Sync', 'Generate']);
  });
  it('should return a titlecase string array for string data types', () => {
    const value = [SupportedJobType.SYNC, SupportedJobType.UNSPECIFIED];
    const result = getTransformerJobTypesString(value);
    expect(result).toStrictEqual(['Sync', 'Unspecified']);
  });
  it('should return empty string if empty array', () => {
    const value: SupportedJobType[] = [];
    const result = getTransformerJobTypesString(value);
    expect(result).toStrictEqual([]);
  });
  it('should return "unspecified" for any invalid values', () => {
    const invalid = 'invalid' as unknown as SupportedJobType;
    const value = [invalid];
    const result = getTransformerJobTypesString(value);
    expect(result).toStrictEqual(['unspecified']);
  });
});

describe('getTransformerSourceString', () => {
  it("should return 'unspecified' for UNSPECIFIED source", () => {
    const input = TransformerSource.UNSPECIFIED;
    const result = getTransformerSourceString(input);
    expect(result).toBe('unspecified');
  });

  it("should return 'passthrough' for PASSTHROUGH source", () => {
    const input = TransformerSource.PASSTHROUGH;
    const result = getTransformerSourceString(input);
    expect(result).toBe('passthrough');
  });

  it("should return 'generate_default' for GENERATE_DEFAULT source", () => {
    const input = TransformerSource.GENERATE_DEFAULT;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_default');
  });

  it("should return 'transform_javascript' for TRANSFORM_JAVASCRIPT source", () => {
    const input = TransformerSource.TRANSFORM_JAVASCRIPT;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_javascript');
  });

  it("should return 'generate_email' for GENERATE_EMAIL source", () => {
    const input = TransformerSource.GENERATE_EMAIL;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_email');
  });

  it("should return 'transform_email' for TRANSFORM_EMAIL source", () => {
    const input = TransformerSource.TRANSFORM_EMAIL;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_email');
  });

  it("should return 'generate_bool' for GENERATE_BOOL source", () => {
    const input = TransformerSource.GENERATE_BOOL;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_bool');
  });

  it("should return 'generate_card_number' for GENERATE_CARD_NUMBER source", () => {
    const input = TransformerSource.GENERATE_CARD_NUMBER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_card_number');
  });

  it("should return 'generate_city' for GENERATE_CITY source", () => {
    const input = TransformerSource.GENERATE_CITY;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_city');
  });

  it("should return 'generate_e164_phone_number' for GENERATE_E164_PHONE_NUMBER source", () => {
    const input = TransformerSource.GENERATE_E164_PHONE_NUMBER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_e164_phone_number');
  });

  it("should return 'generate_first_name' for GENERATE_FIRST_NAME source", () => {
    const input = TransformerSource.GENERATE_FIRST_NAME;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_first_name');
  });

  it("should return 'generate_float64' for GENERATE_FLOAT64 source", () => {
    const input = TransformerSource.GENERATE_FLOAT64;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_float64');
  });

  it("should return 'generate_full_address' for GENERATE_FULL_ADDRESS source", () => {
    const input = TransformerSource.GENERATE_FULL_ADDRESS;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_full_address');
  });

  it("should return 'generate_full_name' for GENERATE_FULL_NAME source", () => {
    const input = TransformerSource.GENERATE_FULL_NAME;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_full_name');
  });

  it("should return 'generate_gender' for GENERATE_GENDER source", () => {
    const input = TransformerSource.GENERATE_GENDER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_gender');
  });

  it("should return 'generate_int64_phone_number' for GENERATE_INT64_PHONE_NUMBER source", () => {
    const input = TransformerSource.GENERATE_INT64_PHONE_NUMBER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_int64_phone_number');
  });

  it("should return 'generate_int64' for GENERATE_INT64 source", () => {
    const input = TransformerSource.GENERATE_INT64;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_int64');
  });

  it("should return 'generate_random_int64' for GENERATE_RANDOM_INT64 source", () => {
    const input = TransformerSource.GENERATE_RANDOM_INT64;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_random_int64');
  });

  it("should return 'generate_last_name' for GENERATE_LAST_NAME source", () => {
    const input = TransformerSource.GENERATE_LAST_NAME;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_last_name');
  });

  it("should return 'generate_sha256hash' for GENERATE_SHA256HASH source", () => {
    const input = TransformerSource.GENERATE_SHA256HASH;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_sha256hash');
  });

  it("should return 'generate_ssn' for GENERATE_SSN source", () => {
    const input = TransformerSource.GENERATE_SSN;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_ssn');
  });

  it("should return 'generate_state' for GENERATE_STATE source", () => {
    const input = TransformerSource.GENERATE_STATE;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_state');
  });

  it("should return 'generate_country' for GENERATE_COUNTRY source", () => {
    const input = TransformerSource.GENERATE_COUNTRY;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_country');
  });

  it("should return 'generate_street_address' for GENERATE_STREET_ADDRESS source", () => {
    const input = TransformerSource.GENERATE_STREET_ADDRESS;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_street_address');
  });

  it("should return 'generate_string_phone_number' for GENERATE_STRING_PHONE_NUMBER source", () => {
    const input = TransformerSource.GENERATE_STRING_PHONE_NUMBER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_string_phone_number');
  });

  it("should return 'generate_string' for GENERATE_STRING source", () => {
    const input = TransformerSource.GENERATE_STRING;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_string');
  });

  it("should return 'generate_random_string' for GENERATE_RANDOM_STRING source", () => {
    const input = TransformerSource.GENERATE_RANDOM_STRING;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_random_string');
  });

  it("should return 'generate_unixtimestamp' for GENERATE_UNIXTIMESTAMP source", () => {
    const input = TransformerSource.GENERATE_UNIXTIMESTAMP;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_unixtimestamp');
  });

  it("should return 'generate_username' for GENERATE_USERNAME source", () => {
    const input = TransformerSource.GENERATE_USERNAME;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_username');
  });

  it("should return 'generate_utctimestamp' for GENERATE_UTCTIMESTAMP source", () => {
    const input = TransformerSource.GENERATE_UTCTIMESTAMP;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_utctimestamp');
  });

  it("should return 'generate_uuid' for GENERATE_UUID source", () => {
    const input = TransformerSource.GENERATE_UUID;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_uuid');
  });

  it("should return 'generate_zipcode' for GENERATE_ZIPCODE source", () => {
    const input = TransformerSource.GENERATE_ZIPCODE;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_zipcode');
  });

  it("should return 'transform_e164_phone_number' for TRANSFORM_E164_PHONE_NUMBER source", () => {
    const input = TransformerSource.TRANSFORM_E164_PHONE_NUMBER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_e164_phone_number');
  });

  it("should return 'transform_first_name' for TRANSFORM_FIRST_NAME source", () => {
    const input = TransformerSource.TRANSFORM_FIRST_NAME;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_first_name');
  });

  it("should return 'transform_float64' for TRANSFORM_FLOAT64 source", () => {
    const input = TransformerSource.TRANSFORM_FLOAT64;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_float64');
  });

  it("should return 'transform_full_name' for TRANSFORM_FULL_NAME source", () => {
    const input = TransformerSource.TRANSFORM_FULL_NAME;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_full_name');
  });

  it("should return 'transform_int64_phone_number' for TRANSFORM_INT64_PHONE_NUMBER source", () => {
    const input = TransformerSource.TRANSFORM_INT64_PHONE_NUMBER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_int64_phone_number');
  });

  it("should return 'transform_int64' for TRANSFORM_INT64 source", () => {
    const input = TransformerSource.TRANSFORM_INT64;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_int64');
  });

  it("should return 'transform_last_name' for TRANSFORM_LAST_NAME source", () => {
    const input = TransformerSource.TRANSFORM_LAST_NAME;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_last_name');
  });

  it("should return 'transform_phone_number' for TRANSFORM_PHONE_NUMBER source", () => {
    const input = TransformerSource.TRANSFORM_PHONE_NUMBER;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_phone_number');
  });

  it("should return 'transform_string' for TRANSFORM_STRING source", () => {
    const input = TransformerSource.TRANSFORM_STRING;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_string');
  });

  it("should return 'generate_null' for GENERATE_NULL source", () => {
    const input = TransformerSource.GENERATE_NULL;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_null');
  });

  it("should return 'generate_categorical' for GENERATE_CATEGORICAL source", () => {
    const input = TransformerSource.GENERATE_CATEGORICAL;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_categorical');
  });

  it("should return 'transform_character_scramble' for TRANSFORM_CHARACTER_SCRAMBLE source", () => {
    const input = TransformerSource.TRANSFORM_CHARACTER_SCRAMBLE;
    const result = getTransformerSourceString(input);
    expect(result).toBe('transform_character_scramble');
  });

  it("should return 'user_defined' for USER_DEFINED source", () => {
    const input = TransformerSource.USER_DEFINED;
    const result = getTransformerSourceString(input);
    expect(result).toBe('user_defined');
  });

  it("should return 'generate_javascript' for GENERATE_JAVASCRIPT source", () => {
    const input = TransformerSource.GENERATE_JAVASCRIPT;
    const result = getTransformerSourceString(input);
    expect(result).toBe('generate_javascript');
  });

  // Test for an invalid source value
  it("should return 'unspecified' for an invalid source", () => {
    const invalidInput = 999; // A number outside the valid range
    const result = getTransformerSourceString(
      invalidInput as TransformerSource
    );
    expect(result).toBe('unspecified');
  });
});

describe('getGenerateEmailTypeString', () => {
  it('should return a string of the full name type', () => {
    const input = GenerateEmailType.FULLNAME;
    const result = getGenerateEmailTypeString(input);
    expect(result).toBe('fullname');
  });
  it('should return a string of the uuid_v4 type', () => {
    const input = GenerateEmailType.UUID_V4;
    const result = getGenerateEmailTypeString(input);
    expect(result).toBe('uuid_v4');
  });
  it('should return a string of the unspecified type', () => {
    const input = GenerateEmailType.UNSPECIFIED;
    const result = getGenerateEmailTypeString(input);
    expect(result).toBe('unspecified');
  });
  it('should return unknown of an invalid type', () => {
    const input = 'test' as unknown as GenerateEmailType.UNSPECIFIED;
    const result = getGenerateEmailTypeString(input);
    expect(result).toBe('unknown');
  });
});

describe('getInvalidEmailActionString', () => {
  it('should return a string of the fgenerate type', () => {
    const input = InvalidEmailAction.GENERATE;
    const result = getInvalidEmailActionString(input);
    expect(result).toBe('generate');
  });
  it('should return a string of the null type', () => {
    const input = InvalidEmailAction.NULL;
    const result = getInvalidEmailActionString(input);
    expect(result).toBe('null');
  });
  it('should return a string of the unspecified type', () => {
    const input = InvalidEmailAction.PASSTHROUGH;
    const result = getInvalidEmailActionString(input);
    expect(result).toBe('passthrough');
  });
  it('should return a string of the reject type', () => {
    const input = InvalidEmailAction.REJECT;
    const result = getInvalidEmailActionString(input);
    expect(result).toBe('reject');
  });
  it('should return a string of the unspecified type', () => {
    const input = InvalidEmailAction.UNSPECIFIED;
    const result = getInvalidEmailActionString(input);
    expect(result).toBe('unspecified');
  });
  it('should return unknown of an invalid type', () => {
    const input = 'test' as unknown as InvalidEmailAction.UNSPECIFIED;
    const result = getInvalidEmailActionString(input);
    expect(result).toBe('unknown');
  });
});
