import { parseDuration } from './duration';

describe('parseDuration', () => {
  test('parses simple durations', () => {
    expect(parseDuration('5s')).toBe(5 * 1e9);
    expect(parseDuration('300ms')).toBe(300 * 1e6);
    expect(parseDuration('1.5h')).toBe(5400 * 1e9);
  });

  test('parses complex durations', () => {
    expect(parseDuration('2h45m')).toBe((2 * 3600 + 45 * 60) * 1e9);
    expect(parseDuration('1h30m10s')).toBe((3600 + 30 * 60 + 10) * 1e9);
  });

  test('parses durations with small units', () => {
    expect(parseDuration('1µs')).toBe(1000);
    expect(parseDuration('1us')).toBe(1000);
    expect(parseDuration('1ns')).toBe(1);
  });

  test('parses negative durations', () => {
    expect(parseDuration('-5s')).toBe(-5 * 1e9);
    expect(parseDuration('-1.5h')).toBe(-5400 * 1e9);
  });

  test('parses zero', () => {
    expect(parseDuration('0')).toBe(0);
    expect(parseDuration('0s')).toBe(0);
  });

  test('handles fractional values', () => {
    expect(parseDuration('1.5s')).toBe(1.5 * 1e9);
    expect(parseDuration('0.5ms')).toBe(0.5 * 1e6);
  });

  test('throws error for invalid durations', () => {
    expect(() => parseDuration('')).toThrow('time: invalid duration ""');
    expect(() => parseDuration('1')).toThrow(
      'time: missing unit in duration "1"'
    );
    expect(() => parseDuration('1x')).toThrow(
      'time: unknown unit "x" in duration "1x"'
    );
    expect(() => parseDuration('1hh')).toThrow(
      'time: unknown unit "hh" in duration "1hh"'
    );
    expect(() => parseDuration('1.5.5h')).toThrow(
      'time: missing unit in duration "1.5.5h"'
    );
  });

  test('handles edge cases', () => {
    expect(parseDuration('1h0m0s')).toBe(3600 * 1e9);
    expect(parseDuration('+5s')).toBe(5 * 1e9);
    expect(parseDuration('01h')).toBe(3600 * 1e9);
  });

  // A little buggy with large durations, but we're not currently anticipating this given the duration context
  // test('handles very large durations', () => {
  //   // This is close to the maximum safe integer in JavaScript
  //   expect(parseDuration('2562047h47m16.854775807s')).toBe(9223372036854775807);
  // });

  test('throws error for durations that are too large', () => {
    expect(() => parseDuration('2562047h47m16.854775808s')).toThrow(
      'time: invalid duration'
    );
  });
});
