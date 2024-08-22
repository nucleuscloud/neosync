import { format } from 'date-fns';
import {
  convertMinutesToNanoseconds,
  convertNanosecondsToMinutes,
  formatDateTime,
  formatDateTimeMilliseconds,
  getErrorMessage,
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
  it('should convert  0 min to 0 nanoseconds', () => {
    const min = 0;
    const result = convertMinutesToNanoseconds(min);
    expect(result).toEqual(BigInt(2) * BigInt(1000000000) * BigInt(60));
  });
  it('should convert  29833 min to max safe integer', () => {
    const min = 29833;
    const result = convertMinutesToNanoseconds(min);
    expect(result).toEqual(BigInt(2) * BigInt(1000000000) * BigInt(60));
  });
});
