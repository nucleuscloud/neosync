import { JsonValue } from '@bufbuild/protobuf';
import { GetSystemTransformerBySourceResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetSystemTransformerBySource(
  source: string
): HookReply<GetSystemTransformerBySourceResponse> {
  return useNucleusAuthenticatedFetch<
    GetSystemTransformerBySourceResponse,
    JsonValue | GetSystemTransformerBySourceResponse
  >(`/api/transformers/system/${source}`, !!source, undefined, (data) =>
    data instanceof GetSystemTransformerBySourceResponse
      ? data
      : GetSystemTransformerBySourceResponse.fromJson(data)
  );
}
