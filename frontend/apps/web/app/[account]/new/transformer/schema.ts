import {
  IsTransformerNameAvailableResponse,
  TransformerConfig,
} from '@neosync/sdk';
import * as Yup from 'yup';

const transformEmailConfig = Yup.object().shape({
  preserveDomain: Yup.boolean().required('This field is required.'),
  preserveLength: Yup.boolean().required('This field is required.'),
});

const generateCardNumberConfig = Yup.object().shape({
  validLuhn: Yup.boolean().required('This field is required.'),
});

const generateE164PhoneNumberConfig = Yup.object().shape({
  min: Yup.number()
    .min(9, 'The value must be greater than or equal to 9.')
    .max(15, 'The value must be less than or equal 15.')
    .required('This field is required.')
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { max } = this.parent;
      return !max || !value || value <= max;
    }),
  max: Yup.number()
    .min(9, 'The value must be greater than or equal to 9.')
    .max(15, 'The value must be less than or equal 15.')
    .required('This field is required.')
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { min } = this.parent;
        return !min || !value || value >= min;
      }
    ),
});
const generateFloat64Config = Yup.object().shape({
  randomizeSign: Yup.bool(),
  min: Yup.number().required('This field is required.'),
  max: Yup.number().max(10).required('This field is required.'),
});

const generateGenderConfig = Yup.object().shape({
  abbreviate: Yup.boolean().required('This field is required.'),
});

const generateInt64Config = Yup.object().shape({
  randomizeSign: Yup.bool().required('This field is required.'),
  min: Yup.number().required('This field is required.'),
  max: Yup.number().required('This field is required.'),
});

const generateStringPhoneNumberConfig = Yup.object().shape({
  includeHyphens: Yup.boolean().required('This field is required.'),
});

const generateStringConfig = Yup.object().shape({
  min: Yup.number()
    .min(1, 'The value must be greater than or equal to 1.')
    .required('This field is required.')
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { max } = this.parent;
      return !max || !value || value <= max;
    }),
  max: Yup.number()
    .min(1, 'The value must be greater than or equal to 1.')
    .required('This field is required.')
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { min } = this.parent;
        return !min || !value || value >= min;
      }
    ),
});

const generateUuidConfig = Yup.object().shape({
  includeHyphens: Yup.boolean().required('This field is required.'),
});

const transformE164PhoneNumber = Yup.object().shape({
  preserveLength: Yup.boolean().required('This field is required.'),
});

const transformFirstNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().required('This field is required.'),
});

const transformFloat64Config = Yup.object().shape({
  randomizationRangeMin: Yup.number()
    .required('This field is required.')
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { randomizationRangeMax } = this.parent;
      return !randomizationRangeMax || !value || value <= randomizationRangeMax;
    }),
  randomizationRangeMax: Yup.number()
    .required('This field is required.')
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { randomizationRangeMin } = this.parent;
        return (
          !randomizationRangeMin || !value || value >= randomizationRangeMin
        );
      }
    ),
});

const transformFullNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().required('This field is required.'),
});

const transformInt64PhoneNumberConfig = Yup.object().shape({
  preserveLength: Yup.boolean().required('This field is required.'),
});

const transformInt64Config = Yup.object().shape({
  randomizationRangeMin: Yup.number()
    .required('This field is required.')
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { randomizationRangeMax } = this.parent;
      return !randomizationRangeMax || !value || value <= randomizationRangeMax;
    }),
  randomizationRangeMax: Yup.number()
    .required('This field is required.')
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { randomizationRangeMin } = this.parent;
        return (
          !randomizationRangeMin || !value || value >= randomizationRangeMin
        );
      }
    ),
});

const transformLastNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean().required('This field is required.'),
});

const transformPhoneNumberConfig = Yup.object().shape({
  preserveLength: Yup.boolean().required('This field is required.'),
  includeHyphens: Yup.boolean().required('This field is required.'),
});

const transformStringConfig = Yup.object().shape({
  preserveLength: Yup.boolean().required('This field is required.'),
});

const userDefinedTransformerConfig = Yup.object().shape({
  id: Yup.string().required('This field is required.'),
});

type ConfigType = TransformerConfig['config'];

// Helper function to extract the 'case' property from a config type
type ExtractCase<T> = T extends { case: infer U } ? U : never;

// Computed type that extracts all case types from the config union
export type TransformerConfigCase = ExtractCase<ConfigType>;

export const EMPTY_TRANSFORMER_CONFIG = Yup.object({
  case: Yup.string(),
  value: Yup.object(),
});

// todo: this should be properly typed
export const TRANSFORMER_SCHEMA_CONFIGS: Record<
  NonNullable<TransformerConfigCase>,
  Yup.ObjectSchema<any> // eslint-disable-line @typescript-eslint/no-explicit-any
> = {
  generateBoolConfig: EMPTY_TRANSFORMER_CONFIG,
  generateCityConfig: EMPTY_TRANSFORMER_CONFIG,
  generateDefaultConfig: EMPTY_TRANSFORMER_CONFIG,
  generateEmailConfig: EMPTY_TRANSFORMER_CONFIG,
  generateFirstNameConfig: EMPTY_TRANSFORMER_CONFIG,
  generateFullAddressConfig: EMPTY_TRANSFORMER_CONFIG,
  generateFullNameConfig: EMPTY_TRANSFORMER_CONFIG,
  generateInt64PhoneNumberConfig: EMPTY_TRANSFORMER_CONFIG,
  generateLastNameConfig: EMPTY_TRANSFORMER_CONFIG,
  generateSha256hashConfig: EMPTY_TRANSFORMER_CONFIG,
  generateSsnConfig: EMPTY_TRANSFORMER_CONFIG,
  generateStateConfig: EMPTY_TRANSFORMER_CONFIG,
  generateStreetAddressConfig: EMPTY_TRANSFORMER_CONFIG,
  generateUnixtimestampConfig: EMPTY_TRANSFORMER_CONFIG,
  generateUsernameConfig: EMPTY_TRANSFORMER_CONFIG,
  generateUtctimestampConfig: EMPTY_TRANSFORMER_CONFIG,
  generateZipcodeConfig: EMPTY_TRANSFORMER_CONFIG,
  nullconfig: EMPTY_TRANSFORMER_CONFIG,
  passthroughConfig: EMPTY_TRANSFORMER_CONFIG,

  transformEmailConfig: transformEmailConfig,
  generateCardNumberConfig: generateCardNumberConfig,
  generateE164PhoneNumberConfig: generateE164PhoneNumberConfig,
  generateFloat64Config: generateFloat64Config,
  generateGenderConfig: generateGenderConfig,
  generateInt64Config: generateInt64Config,
  generateStringPhoneNumberConfig: generateStringPhoneNumberConfig,
  generateStringConfig: generateStringConfig,
  generateUuidConfig: generateUuidConfig,
  transformE164PhoneNumberConfig: transformE164PhoneNumber,
  transformFirstNameConfig: transformFirstNameConfig,
  transformFloat64Config: transformFloat64Config,
  transformFullNameConfig: transformFullNameConfig,
  transformInt64PhoneNumberConfig: transformInt64PhoneNumberConfig,
  transformInt64Config: transformInt64Config,
  transformLastNameConfig: transformLastNameConfig,
  transformPhoneNumberConfig: transformPhoneNumberConfig,
  transformStringConfig: transformStringConfig,
  userDefinedTransformerConfig: userDefinedTransformerConfig,
};

export const TransformerConfigSchema = Yup.lazy((v) => {
  const ccase = v?.case as TransformerConfigCase;
  if (!ccase) {
    return EMPTY_TRANSFORMER_CONFIG;
  }
  const cconfig = TRANSFORMER_SCHEMA_CONFIGS[ccase];
  return Yup.object({
    case: Yup.string().required().oneOf([ccase]),
    value: cconfig,
  });
});

// Simplified version of a job mapping transformer config for use with react-hook-form only
export type TransformerConfigSchema = Yup.InferType<
  typeof TransformerConfigSchema
>;

const transformerNameSchema = Yup.string()
  .required()
  .test(
    'checkNameUnique',
    'Transformer Name must be at least 3 characters long and can only include lowercase letters, numbers, and hyphens.',
    async (value, context) => {
      if (!value || value.length < 3) {
        return context.createError({
          message:
            'Transformer name is too short. Must be at least 3 characters long.',
        });
      }

      const regex = /^[a-z0-9-]+$/;
      if (!regex.test(value)) {
        return context.createError({
          message:
            'Transformer name Name can only include lowercase letters, numbers, and hyphens.',
        });
      }
      const accountId = context?.options?.context?.accountId;
      if (!accountId) {
        return context.createError({
          message: 'Unable to verify Name. Account is not valid.',
        });
      }

      if (value == context?.options?.context?.name) {
        return true;
      }

      try {
        const res = await isTransformerNameAvailable(value, accountId);
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
  config: TransformerConfigSchema,
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
  config: TransformerConfigSchema,
});

export type UpdateUserDefinedTransformer = Yup.InferType<
  typeof UPDATE_USER_DEFINED_TRANSFORMER
>;
async function isTransformerNameAvailable(
  name: string,
  accountId: string
): Promise<IsTransformerNameAvailableResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/transformers/is-transformer-name-available?transformerName=${name}`,
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
  config: TransformerConfigSchema,
});

export type SystemTransformersSchema = Yup.InferType<
  typeof SYSTEM_TRANSFORMER_SCHEMA
>;
