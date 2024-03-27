import {
  filterInputFreeSystemTransformers,
  filterInputFreeUdTransformers,
} from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { SystemTransformer, UserDefinedTransformer } from '@neosync/sdk';
import { useMemo } from 'react';
import { useGetSystemTransformers } from './useGetSystemTransformers';
import { useGetUserDefinedTransformers } from './useGetUserDefinedTransformers';

interface GetMergedTransformersResponse {
  systemTransformers: SystemTransformer[];
  systemMap: Map<string, SystemTransformer>;
  userDefinedTransformers: UserDefinedTransformer[];
  userDefinedMap: Map<string, UserDefinedTransformer>;
  isLoading: boolean;
  isValidating: boolean;
}

export function useGetMergedTransformers(
  excludeInputReqTransformers: boolean,
  accountId: string
): GetMergedTransformersResponse {
  const {
    data: systemTransformersData,
    isLoading: isLoadingSystemTransformers,
    isValidating: isSystemValidating,
  } = useGetSystemTransformers();
  const {
    data: customTransformersData,
    isLoading: isLoadingCustomTransformers,
    isValidating: isCustomValidating,
  } = useGetUserDefinedTransformers(accountId);

  const isLoading = isLoadingSystemTransformers || isLoadingCustomTransformers;
  const isValidating = isSystemValidating || isCustomValidating;

  const { filteredSystem, filteredCustom, userMap, sysMap } = useMemo(() => {
    const systemTransformers = systemTransformersData?.transformers ?? [];
    const customTransformers = customTransformersData?.transformers ?? [];

    const filteredSystemTransformers = excludeInputReqTransformers
      ? filterInputFreeSystemTransformers(systemTransformers)
      : systemTransformers;

    const filteredCustomTransformers = excludeInputReqTransformers
      ? filterInputFreeUdTransformers(
          customTransformers,
          filteredSystemTransformers
        )
      : customTransformers;

    return {
      filteredSystem: filteredSystemTransformers,
      filteredCustom: filteredCustomTransformers,
      sysMap: new Map(filteredSystemTransformers.map((t) => [t.source, t])),
      userMap: new Map(filteredCustomTransformers.map((t) => [t.id, t])),
    };
  }, [isValidating, excludeInputReqTransformers]);

  return {
    systemTransformers: filteredSystem,
    userDefinedTransformers: filteredCustom,
    systemMap: sysMap,
    userDefinedMap: userMap,

    isLoading: isLoading,
    isValidating: isValidating,
  };
}
