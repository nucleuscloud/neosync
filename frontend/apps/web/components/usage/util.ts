import { create } from '@bufbuild/protobuf';
import {
  Date as NeosyncDate,
  DateSchema as NeosyncDateSchema,
} from '@neosync/sdk';
import { endOfMonth, format, startOfMonth, subMonths } from 'date-fns';

export type MetricIdentifierType = 'accountId' | 'jobId' | 'runId';

export function shortNumberFormatter(
  formatter: Intl.NumberFormat,
  value: number
): string {
  if (Math.abs(value) >= 1_000_000_000_000) {
    return formatter.format(value / 1_000_000_000_000) + 'T';
  } else if (Math.abs(value) >= 1_000_000_000) {
    return formatter.format(value / 1_000_000_000) + 'B';
  } else if (Math.abs(value) >= 1_000_000) {
    return formatter.format(value / 1_000_000) + 'M';
  } else if (Math.abs(value) >= 1_000) {
    return formatter.format(value / 1_000) + 'K';
  } else {
    return formatter.format(value);
  }
}

export type UsagePeriod = 'current' | 'last-month';

export function dateToNeoDate(date: Date): NeosyncDate {
  return create(NeosyncDateSchema, {
    day: date.getDate(),
    month: date.getMonth() + 1,
    year: date.getFullYear(),
  });
}

export function periodToDateRange(period: UsagePeriod): [Date, Date] {
  const currentDate = new Date();
  switch (period) {
    case 'current': {
      const start = startOfMonth(currentDate);
      const end = endOfMonth(currentDate);
      return [start, end];
    }
    case 'last-month': {
      const prevMonthDate = subMonths(currentDate, 1);
      const start = startOfMonth(prevMonthDate);
      const end = endOfMonth(prevMonthDate);
      return [start, end];
    }
  }
}

export function getPeriodLabel(period: UsagePeriod): string {
  switch (period) {
    case 'current': {
      return 'Current Period';
    }
    case 'last-month': {
      return 'Last Month';
    }
  }
}

export function getDateRangeLabel(start: Date, end: Date): string {
  return `${format(start, 'MM/dd/yy')} - ${format(end, 'MM/dd/yy')}`;
}
