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

const userDefinedTransformerConfig = Yup.object().shape({
  id: Yup.string().required(),
});

export const transformerConfig = Yup.object().shape({
  config: Yup.object().shape({
    value: Yup.lazy((value) => {
      switch (value?.case) {
        case 'generateEmailConfig':
          return Yup.object().shape({});
        case 'generateRealisticEmailConfig':
          return Yup.object().shape({});
        case 'transformEmailConfig':
          return transformEmailConfig;
        case 'generateBoolConfig':
          return Yup.object().shape({});
        case 'generateCardNumberConfig':
          return generateCardNumberConfig;
        case 'generateCityConfig':
          return Yup.object().shape({});
        case 'generateE164NumberConfig':
          return generateE164NumberConfig;
        case 'generateFirstNameConfig':
          return Yup.object().shape({});
        case 'generateFloatConfig':
          return generateFloatConfig;
        case 'generateFullAddressConfig':
          return Yup.object().shape({});
        case 'generateFullNameConfig':
          return Yup.object().shape({});
        case 'generateGenderConfig':
          return generateGenderConfig;
        case 'generateInt64PhoneConfig':
          return Yup.object().shape({});
        case 'generateIntConfig':
          return generateIntConfig;
        case 'generateLastNameConfig':
          return Yup.object().shape({});
        case 'generateSha256hashConfig':
          return Yup.object().shape({});
        case 'generateSsnConfig':
          return Yup.object().shape({});
        case 'generateStateConfig':
          return Yup.object().shape({});
        case 'generateStreetAddressConfig':
          return Yup.object().shape({});
        case 'generateStringPhoneConfig':
          return generateStringPhoneConfig;
        case 'generateStringConfig':
          return generateStringConfig;
        case 'generateUnixtimestampConfig':
          return Yup.object().shape({});
        case 'generateUsernameConfig':
          return Yup.object().shape({});
        case 'generateUtctimestampConfig':
          return Yup.object().shape({});
        case 'generateUuidConfig':
          return generateUuidConfig;
        case 'generateZipcodeConfig':
          return Yup.object().shape({});
        case 'transformE164PhoneConfig':
          return transformE164Phone;
        case 'transformFirstNameConfig':
          return transformFirstNameConfig;
        case 'transformFloatConfig':
          return transformFloatConfig;
        case 'transformFullNameConfig':
          return transformFullNameConfig;
        case 'transformIntPhoneConfig':
          return transformIntPhoneConfig;
        case 'transformIntConfig':
          return transformIntConfig;
        case 'transformLastNameConfig':
          return transformLastNameConfig;
        case 'transformPhoneConfig':
          return transformPhoneConfig;
        case 'transformStringConfig':
          return transformStringConfig;
        case 'passthroughConfig':
          return Yup.object().shape({});
        case 'userDefinedTransformerConfig':
          return userDefinedTransformerConfig;
        case 'nullconfig':
          return Yup.object().shape({});
        default:
          return Yup.object().shape({});
      }
    }),
    case: Yup.string(),
  }),
});

const transformerNameSchema = Yup.string()
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

export const CREATE_USER_DEFINED_TRANSFORMER_SCHEMA = Yup.object({
  name: transformerNameSchema,
  source: Yup.string(),
  description: Yup.string().required(),
  type: Yup.string().required(),
  config: transformerConfig,
});

export type CreateUserDefinedTransformerSchema = Yup.InferType<
  typeof CREATE_USER_DEFINED_TRANSFORMER_SCHEMA
>;

export const UPDATE_USER_DEFINED_TRANSFORMER = Yup.object({
  name: transformerNameSchema,
  id: Yup.string(),
  source: Yup.string(),
  description: Yup.string().required(),
  type: Yup.string(),
  config: transformerConfig,
});

export type UpdateUserDefinedTransformer = Yup.InferType<
  typeof UPDATE_USER_DEFINED_TRANSFORMER
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
