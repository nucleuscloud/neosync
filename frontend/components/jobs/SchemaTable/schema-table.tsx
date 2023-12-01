import {
  MergeSystemAndCustomTransformers,
  filterDataTransformers,
} from '@/app/transformers/EditTransformerOptions';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetCustomTransformers } from '@/libs/hooks/useGetCustomTransformers';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { GetConnectionSchemaResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { ReactElement } from 'react';
import { VirtualizedSchemaTable } from './VirtualizedSchemaTable';

interface Props {
  data?: JobMappingFormValues[];
  excludeTransformers?: boolean; // will result in only generators (functions with no data input)
}

export function SchemaTable(props: Props): ReactElement {
  const { data, excludeTransformers } = props;

  const { account } = useAccount();
  const { data: systemTransformers, isLoading: systemTransformersIsLoading } =
    useGetSystemTransformers();

  const { data: customTransformers, isLoading: customTransformersIsLoading } =
    useGetCustomTransformers(account?.id ?? '');

  const filteredSystemTransformers = excludeTransformers
    ? filterDataTransformers(systemTransformers?.transformers ?? [])
    : systemTransformers?.transformers ?? [];

  const mergedTransformers = MergeSystemAndCustomTransformers(
    filteredSystemTransformers,
    customTransformers?.transformers ?? []
  );

  console.log('table data', data);

  const tableData = data?.map((d) => {
    return {
      ...d,
      isSelected: false,
    };
  });

  if (
    systemTransformersIsLoading ||
    customTransformersIsLoading ||
    !tableData
  ) {
    return <SkeletonTable />;
  }

  return (
    <div>
      <VirtualizedSchemaTable
        data={tableData}
        transformers={mergedTransformers}
      />
    </div>
  );
}

export async function getConnectionSchema(
  connectionId?: string
): Promise<GetConnectionSchemaResponse | undefined> {
  if (!connectionId) {
    return;
  }
  const res = await fetch(`/api/connections/${connectionId}/schema`, {
    method: 'GET',
    headers: {
      'content-type': 'application/json',
    },
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionSchemaResponse.fromJson(await res.json());
}

export interface TransformerWithType extends Transformer {
  transformerType: 'system' | 'custom';
}

export function MergeSystemAndCustomTransformers(
  system: Transformer[],
  custom: Transformer[]
): TransformerWithType[] {
  const newSystem = system.map((item) => ({
    ...item,
    transformerType: 'system',
  }));

  const newCustom = custom.map((item) => ({
    ...item,
    transformerType: 'custom',
  }));

  const combinedArray = [...newSystem, ...newCustom];

  return combinedArray as TransformerWithType[];
}
