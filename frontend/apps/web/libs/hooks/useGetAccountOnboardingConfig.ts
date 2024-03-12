import { JsonValue } from '@bufbuild/protobuf';
import { GetAccountOnboardingConfigResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountOnboardingConfig(
  accountId: string
): HookReply<GetAccountOnboardingConfigResponse> {
  return useNucleusAuthenticatedFetch<
    GetAccountOnboardingConfigResponse,
    JsonValue | GetAccountOnboardingConfigResponse
  >(
    `/api/users/accounts/${accountId}/onboarding-config`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetAccountOnboardingConfigResponse
        ? data
        : GetAccountOnboardingConfigResponse.fromJson(data)
  );
}
