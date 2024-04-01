import { JsonValue } from '@bufbuild/protobuf';
import {
  GetSystemTransformerBySourceResponse,
  TransformerSource,
} from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetSystemTransformerBySource(
  source: TransformerSource
): HookReply<GetSystemTransformerBySourceResponse> {
  return useNucleusAuthenticatedFetch<
    GetSystemTransformerBySourceResponse,
    JsonValue | GetSystemTransformerBySourceResponse
  >(`/api/transformers/system/${source}`, source != null, undefined, (data) =>
    data instanceof GetSystemTransformerBySourceResponse
      ? data
      : GetSystemTransformerBySourceResponse.fromJson(data)
  );
}
