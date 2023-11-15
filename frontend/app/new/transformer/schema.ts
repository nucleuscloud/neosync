import { getAccount } from '@/components/providers/account-provider';
import { IsTransformerNameAvailableResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import * as Yup from 'yup';

const emailConfig = Yup.object().shape({
  preserveDomain: Yup.boolean().notRequired(),
  preserveLength: Yup.boolean().notRequired(),
});

const uuidConfig = Yup.object().shape({
  includeHyphens: Yup.boolean().notRequired(),
});

const firstNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const lastNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const fullNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const phoneNumberConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  e164Format: Yup.boolean().notRequired(),
  includeHyphens: Yup.boolean().notRequired(),
});

const intPhoneNumberConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const randomStringConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  strLength: Yup.number().notRequired(),
});
const randomInt = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  intLength: Yup.number().notRequired(),
});
const genderConfig = Yup.object().shape({
  abbreviate: Yup.boolean().notRequired(),
});

const randomFloatConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  digitsAfterDecimal: Yup.number().notRequired(),
  digitsBeforeDecimal: Yup.number().notRequired(),
});

const cardNumberConfig = Yup.object().shape({
  validLuhn: Yup.boolean().notRequired(),
});

export const transformerConfig = Yup.object().shape({
  config: Yup.object().shape({
    value: Yup.lazy((value) => {
      switch (value?.case) {
        case 'emailConfig':
          return emailConfig;
        case 'passthroughConfig':
          return Yup.object().shape({});
        case 'uuidConfig':
          return uuidConfig;
        case 'firstNameConfig':
          return firstNameConfig;
        case 'lastNameConfig':
          return lastNameConfig;
        case 'fullNameConfig':
          return fullNameConfig;
        case 'phoneNumberConfig':
          return phoneNumberConfig;
        case 'intPhoneNumberConfig':
          return intPhoneNumberConfig;
        case 'nullConfig':
          return Yup.object().shape({});
        case 'randomStringConfig':
          return randomStringConfig;
        case 'randomBoolConfig':
          return Yup.object().shape({});
        case 'nullConfig':
          return Yup.object().shape({});
        case 'randomInt':
          return randomInt;
        case 'gender':
          return genderConfig;
        case 'randomFloatConfig':
          return randomFloatConfig;
        case 'utcTimestampConfig':
          return Yup.object().shape({});
        case 'unix_timestamp':
          return Yup.object().shape({});
        case 'cityConfig':
          return Yup.object().shape({});
        case 'zipcodeConfig':
          return Yup.object().shape({});
        case 'stateConfig':
          return Yup.object().shape({});
        case 'fullAddressConfig':
          return Yup.object().shape({});
        case 'streetAddressConfig':
          return Yup.object().shape({});
        case 'cardNumberConfig':
          return cardNumberConfig;
        case 'sha256hashConfig':
          return Yup.object().shape({});
        case 'ssnConfig':
          return Yup.object().shape({});
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

export const SYSTEM_TRANSFORMER_SCHEMA = Yup.object({
  name: Yup.string(),
  type: Yup.string(),
  description: Yup.string().required(),
  config: transformerConfig,
});

export type SystemTransformersSchema = Yup.InferType<
  typeof SYSTEM_TRANSFORMER_SCHEMA
>;
