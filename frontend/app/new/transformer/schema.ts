import { getAccount } from '@/components/providers/account-provider';
import { IsTransformerNameAvailableResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import * as Yup from 'yup';

export const DEFINE_NEW_TRANSFORMER_SCHEMA = Yup.object({
  transformerName: Yup.string()
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
  baseTransformer: Yup.string(),
});

export type DefineNewTransformer = Yup.InferType<
  typeof DEFINE_NEW_TRANSFORMER_SCHEMA
>;

async function isTransformerNameAvailable(
  name: string,
  accountId: string
): Promise<IsTransformerNameAvailableResponse> {
  const res = await fetch(
    `/api/transformer/is-transformer-name-available?transformerName=${name}&accountId=${accountId}`,
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
