'use client';
import { JsonValue } from '@bufbuild/protobuf';
import {
  GetDailyMetricCountResponse,
  Date as NeosyncDate,
  RangedMetricName,
} from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

type MetricIdentifierType = 'accountId' | 'jobId' | 'runId';

export function useGetDailyMetricCount(
  accountId: string,
  start: NeosyncDate,
  end: NeosyncDate,
  metric: RangedMetricName,
  idtype: MetricIdentifierType,
  identifier: string
): HookReply<GetDailyMetricCountResponse> {
  const urlparams = new URLSearchParams({
    startDay: start.day.toString(),
    startMo: start.month.toString(),
    startYear: start.year.toString(),
    endDay: end.day.toString(),
    endMo: end.month.toString(),
    endYear: end.year.toString(),
    metric: metric.toString(),
    idtype,
    identifier,
  });
  return useNucleusAuthenticatedFetch<
    GetDailyMetricCountResponse,
    JsonValue | GetDailyMetricCountResponse
  >(
    `/api/accounts/${accountId}/metrics/daily-count?${urlparams.toString()}`,
    !!accountId && !!start && !!end && !!metric && !!idtype && !!identifier,
    undefined,
    (data) =>
      data instanceof GetDailyMetricCountResponse
        ? data
        : GetDailyMetricCountResponse.fromJson(data)
  );
}
