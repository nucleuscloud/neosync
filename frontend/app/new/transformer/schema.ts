import { getAccount } from '@/components/providers/account-provider';
import { IsTransformerNameAvailableResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import * as Yup from 'yup';

const emailConfig = Yup.object().shape({
  preserve_domain: Yup.boolean().notRequired(),
  preserve_length: Yup.boolean().notRequired(),
});

export const transformerConfig = Yup.object().shape({
  config: Yup.object().shape({
    value: Yup.lazy((value) => {
      switch (value?.case) {
        case 'emailConfig':
          return emailConfig;
        default:
          return Yup.object().shape({});
      }
    }),
    case: Yup.string(),
  }),
});

export const CREATE_CUSTOM_TRANSFORMER_SCHEMA = Yup.object({
  name: Yup.string()
    .trim()
    .required('Name is a required field')
    .min(3)
    .max(30)
    .test('checkNameUnique', 'This name is already taken.', async (value) => {
      if (!value || value.length == 0) {
        return false;
      }
      const account = getAccount();
      if (!account) {
        return false;
      }
      const res = await isTransformerNameAvailable(value, account.id);
      return res.isAvailable;
    }),
  source: Yup.string(),
  description: Yup.string().required(),
  config: transformerConfig,
});

export type CreateCustomTransformerSchema = Yup.InferType<
  typeof CREATE_CUSTOM_TRANSFORMER_SCHEMA
>;

export const UPDATE_CUSTOM_TRANSFORMER = Yup.object({
  name: Yup.string()
    .trim()
    .required('Name is a required field')
    .min(3)
    .max(30)
    .test('checkNameUnique', 'This name is already taken.', async (value) => {
      if (!value || value.length == 0) {
        return false;
      }
      const account = getAccount();
      if (!account) {
        return false;
      }
      const res = await isTransformerNameAvailable(value, account.id);
      return res.isAvailable;
    }),
  id: Yup.string(),
  source: Yup.string(),
  description: Yup.string().required(),
  type: Yup.string().required(),
  config: transformerConfig,
});

export type UpdateCustomTransformer = Yup.InferType<
  typeof UPDATE_CUSTOM_TRANSFORMER
>;
async function isTransformerNameAvailable(
  name: string,
  accountId: string
): Promise<IsTransformerNameAvailableResponse> {
  const res = await fetch(
    `/api/transformers/is-transformer-name-available?transformerName=${name}&accountId=${accountId}`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return IsTransformerNameAvailableResponse.fromJson(await res.json());
}
