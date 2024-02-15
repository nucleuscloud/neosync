import {
  filterInputFreeSystemTransformers,
  filterInputFreeUdfTransformers,
} from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { Transformer, joinTransformers } from '@/shared/transformers';
import { useMemo } from 'react';
import { useGetSystemTransformers } from './useGetSystemTransformers';
import { useGetUserDefinedTransformers } from './useGetUserDefinedTransformers';

interface GetMergedTransformersResponse {
  mergedTransformers: Transformer[];
  isLoading: boolean;
}

export function useGetMergedTransformers(
  excludeInputReqTransformers: boolean,
  accountId: string
): GetMergedTransformersResponse {
  const { data: systemTransformers, isLoading: isLoadingSystemTransformers } =
    useGetSystemTransformers();
  const { data: customTransformers, isLoading: isLoadingCustomTransformers } =
    useGetUserDefinedTransformers(accountId);

  const isLoading = isLoadingSystemTransformers || isLoadingCustomTransformers;

  const mergedTransformers = useMemo(() => {
    if (!systemTransformers || !customTransformers) return [];

    const filteredSystemTransformers = excludeInputReqTransformers
      ? filterInputFreeSystemTransformers(systemTransformers.transformers ?? [])
      : systemTransformers.transformers ?? [];

    const filteredCustomTransformers = excludeInputReqTransformers
      ? filterInputFreeUdfTransformers(
          customTransformers.transformers ?? [],
          filteredSystemTransformers
        )
      : customTransformers.transformers ?? [];

    return joinTransformers(
      filteredSystemTransformers,
      filteredCustomTransformers
    );
  }, [systemTransformers, customTransformers, excludeInputReqTransformers]);

  const res = {
    mergedTransformers: mergedTransformers,
    isLoading: isLoading,
  };

  return res;
}
