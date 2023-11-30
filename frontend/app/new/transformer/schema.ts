import { getAccount } from '@/components/providers/account-provider';
import { IsTransformerNameAvailableResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import * as Yup from 'yup';

const transformEmailConfig = Yup.object().shape({
  preserveDomain: Yup.boolean().notRequired(),
  preserveLength: Yup.boolean().notRequired(),
});

const generateCardNumberConfig = Yup.object().shape({
  validLuhn: Yup.boolean().notRequired(),
});

const generateE164NumberConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  e164Format: Yup.boolean().notRequired(),
  includeHyphens: Yup.boolean().notRequired(),
});

const generateFloatConfig = Yup.object().shape({
  sign: Yup.string().notRequired(),
  digitsAfterDecimal: Yup.number().notRequired(),
  digitsBeforeDecimal: Yup.number().notRequired(),
});

const generateGenderConfig = Yup.object().shape({
  abbreviate: Yup.boolean().notRequired(),
});

const generateIntConfig = Yup.object().shape({
  sign: Yup.string().notRequired(),
  length: Yup.number().notRequired(),
});

const generateStringPhoneConfig = Yup.object().shape({
  e164Format: Yup.boolean().notRequired(),
  includeHyphens: Yup.boolean().notRequired(),
});

const generateStringConfig = Yup.object().shape({
  length: Yup.number().notRequired(),
});

const generateUuidConfig = Yup.object().shape({
  includeHyphens: Yup.boolean().notRequired(),
});

const transformE164Phone = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const transformFirstNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const transformFloatConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  preserveSign: Yup.string().notRequired(),
});

const transformFullNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const transformIntPhoneConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const transformIntConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  preserveSign: Yup.boolean().notRequired(),
});

const transformLastNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const transformPhoneConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
  includeHyphens: Yup.boolean().notRequired(),
});

const transformStringConfig = Yup.object().shape({
  preserveLength: Yup.boolean().notRequired(),
});

const customTransformerConfig = Yup.object().shape({
  id: Yup.string().required(),
});

export const transformerConfig = Yup.object().shape({
  config: Yup.object().shape({
    value: Yup.lazy((value) => {
      switch (value?.case) {
        case 'generate_email':
          return Yup.object().shape({});
        case 'generate_realistic_email':
          return Yup.object().shape({});
        case 'transform_email':
          return transformEmailConfig;
        case 'generate_bool':
          return Yup.object().shape({});
        case 'generate_card_number':
          return generateCardNumberConfig;
        case 'generate_city':
          return Yup.object().shape({});
        case 'generate_e164_number':
          return generateE164NumberConfig;
        case 'generate_first_name':
          return Yup.object().shape({});
        case 'generate_float':
          return generateFloatConfig;
        case 'generate_full_address':
          return Yup.object().shape({});
        case 'generate_full_name':
          return Yup.object().shape({});
        case 'generate_gender':
          return generateGenderConfig;
        case 'generate_int64_phone':
          return Yup.object().shape({});
        case 'generate_int':
          return generateIntConfig;
        case 'generate_last_name':
          return Yup.object().shape({});
        case 'generate_sha256hash':
          return Yup.object().shape({});
        case 'generate_ssn':
          return Yup.object().shape({});
        case 'generate_state':
          return Yup.object().shape({});
        case 'generate_street_address':
          return Yup.object().shape({});
        case 'generate_string_phone':
          return generateStringPhoneConfig;
        case 'generate_string':
          return generateStringConfig;
        case 'generate_unixtimestamp':
          return Yup.object().shape({});
        case 'generate_username':
          return Yup.object().shape({});
        case 'generate_utctimestamp':
          return Yup.object().shape({});
        case 'generate_uuid':
          return generateUuidConfig;
        case 'generate_zipcode':
          return Yup.object().shape({});
        case 'transform_e164_phone':
          return transformE164Phone;
        case 'transform_first_name':
          return transformFirstNameConfig;
        case 'transform_float':
          return transformFloatConfig;
        case 'transform_full_name':
          return transformFullNameConfig;
        case 'transform_int_phone':
          return transformIntPhoneConfig;
        case 'transform_int':
          return transformIntConfig;
        case 'transform_last_name':
          return transformLastNameConfig;
        case 'transform_phone':
          return transformPhoneConfig;
        case 'transform_string':
          return transformStringConfig;
        case 'passthrough':
          return Yup.object().shape({});
        case 'custom_transformer_config':
          return customTransformerConfig;
        case 'null':
          return Yup.object().shape({});
        default:
          return Yup.object().shape({});
      }
    }),
    case: Yup.string(),
  }),
});
const customTransformerNameSchema = Yup.string()
  .required()
  .test(
    'checkNameUnique',
    'Transformer Name must be at least 3 characters long and can only include lowercase letters, numbers, and hyphens.',
    async (value, context) => {
      if (!value || value.length < 3) {
        return context.createError({
          message:
            'Transformer is too short. Must be at least 3 characters long.',
        });
      }

      const regex = /^[a-z0-9-]+$/;
      if (!regex.test(value)) {
        return context.createError({
          message:
            'Transformer name Name can only include lowercase letters, numbers, and hyphens.',
        });
      }

      const account = getAccount();
      if (!account) {
        return context.createError({
          message: 'Account is not valid.',
        });
      }

      if (value == context?.options?.context?.name) {
        return true;
      }

      try {
        const res = await isTransformerNameAvailable(value, account.id);
        if (!res.isAvailable) {
          return context.createError({
            message: 'This Transformer Name is already taken.',
          });
        }
        return true;
      } catch (error) {
        return context.createError({
          message: 'Error validating name availability.',
        });
      }
    }
  );

export const CREATE_CUSTOM_TRANSFORMER_SCHEMA = Yup.object({
  name: customTransformerNameSchema,
  source: Yup.string(),
  description: Yup.string().required(),
  config: transformerConfig,
});

export type CreateCustomTransformerSchema = Yup.InferType<
  typeof CREATE_CUSTOM_TRANSFORMER_SCHEMA
>;

export const UPDATE_CUSTOM_TRANSFORMER = Yup.object({
  name: customTransformerNameSchema,
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
