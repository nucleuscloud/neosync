'use client';
import { JsonValue } from '@bufbuild/protobuf';
import {
  GetMetricCountResponse,
  Date as NeosyncDate,
  RangedMetricName,
} from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export type MetricIdentifierType = 'accountId' | 'jobId' | 'runId';

export function useGetMetricCount(
  accountId: string,
  start: NeosyncDate,
  end: NeosyncDate,
  metric: RangedMetricName,
  idtype: MetricIdentifierType,
  identifier: string
): HookReply<GetMetricCountResponse> {
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
    GetMetricCountResponse,
    JsonValue | GetMetricCountResponse
  >(
    `/api/accounts/${accountId}/metrics/count?${urlparams.toString()}`,
    !!accountId && !!start && !!end && !!metric && !!idtype && !!identifier,
    undefined,
    (data) =>
      data instanceof GetMetricCountResponse
        ? data
        : GetMetricCountResponse.fromJson(data)
  );
}
