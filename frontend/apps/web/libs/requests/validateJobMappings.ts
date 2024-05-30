import {
  JobMappingFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
} from '@/yup-validations/jobs';
import {
  JobMapping,
  JobMappingTransformer,
  TransformerConfig,
  ValidateJobMappingsRequest,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';

export async function validateJobMapping(
  connectionId: string,
  formMappings: JobMappingFormValues[],
  accountId: string
): Promise<ValidateJobMappingsResponse> {
  const body = new ValidateJobMappingsRequest({
    accountId,
    mappings: formMappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer:
          m.transformer.source != 0
            ? convertJobMappingTransformerFormToJobMappingTransformer(
                m.transformer
              )
            : new JobMappingTransformer({
                source: 1,
                config: new TransformerConfig({
                  config: {
                    case: 'passthroughConfig',
                    value: {},
                  },
                }),
              }),
      });
    }),
    connectionId: connectionId,
  });

  const res = await fetch(`/api/accounts/${accountId}/jobs/validate-mappings`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }

  return ValidateJobMappingsResponse.fromJson(await res.json());
}
