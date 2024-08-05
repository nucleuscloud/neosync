import { RESOURCE_NAME_REGEX } from '@/yup-validations/connections';
import { PartialMessage } from '@bufbuild/protobuf';
import {
  ConnectError,
  IsTransformerNameAvailableRequest,
  IsTransformerNameAvailableResponse,
  TransformerConfig,
  ValidateUserJavascriptCodeRequest,
  ValidateUserJavascriptCodeResponse,
} from '@neosync/sdk';
import { UseMutateAsyncFunction } from '@tanstack/react-query';
import * as Yup from 'yup';
import { tryBigInt } from '../../transformers/Sheetforms/util';

const bigIntValidator = Yup.mixed<bigint>().test(
  'is-bigint',
  'Value must be bigint',
  (value) => {
    if (typeof value === 'bigint') {
      return true;
    } else if (typeof value === 'number') {
      return true;
    } else if (typeof value === 'string') {
      try {
        BigInt(value);
        return true;
      } catch {
        return false;
      }
    }
  }
);

function getBigIntMinValidator(
  minVal: number | string | bigint
): (value: bigint | undefined) => boolean {
  return (value) => {
    if (value === undefined || value === null) {
      return false;
    }
    const MIN_VALUE = BigInt(minVal);
    try {
      const bigIntValue = BigInt(value);
      return bigIntValue >= MIN_VALUE;
    } catch {
      return false; // Not convertible to BigInt, but this should theoretically not happen due to previous test
    }
  };
}
function getBigIntMaxValidator(
  maxVal: number | string | bigint
): (value: bigint | undefined) => boolean {
  return (value) => {
    if (value === undefined || value === null) {
      return false;
    }
    const MAX_VALUE = BigInt(maxVal);
    try {
      const bigIntValue = BigInt(value);
      return bigIntValue <= MAX_VALUE;
    } catch {
      return false; // Not convertible to BigInt, but this should theoretically not happen due to previous test
    }
  };
}

const transformEmailConfig = Yup.object().shape({
  preserveDomain: Yup.boolean()
    .default(false)
    .required('This field is required.'),
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
  excludedDomains: Yup.array()
    .of(Yup.string().required())
    .optional()
    .default([]),
  emailType: Yup.string().default('GENERATE_EMAIL_TYPE_UUID_V4'),
  invalidEmailAction: Yup.string().default('INVALID_EMAIL_ACTION_REJECT'),
});

const generateEmailConfig = Yup.object().shape({
  emailType: Yup.string().default('GENERATE_EMAIL_TYPE_UUID_V4'),
});

const generateCardNumberConfig = Yup.object().shape({
  validLuhn: Yup.boolean().default(false).required('This field is required.'),
});

const generateInternationalPhoneNumberConfig = Yup.object().shape({
  min: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 9',
      getBigIntMinValidator(9)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${15}`,
      getBigIntMaxValidator(15)
    )
    .required('This field is required.')
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { max } = this.parent;
      const maxBig = tryBigInt(max);
      const valueBig = tryBigInt(value);
      return maxBig !== null && valueBig !== null && valueBig <= maxBig;
    }),
  max: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 9',
      getBigIntMinValidator(9)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${15}`,
      getBigIntMaxValidator(15)
    )
    .required('This field is required.')
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { min } = this.parent;
        const valueBig = tryBigInt(value);
        const minBig = tryBigInt(min);
        return minBig !== null && valueBig !== null && valueBig >= minBig;
      }
    ),
});
const generateFloat64Config = Yup.object().shape({
  randomizeSign: Yup.bool().default(false),
  min: Yup.number()
    .required('This field is required.')
    .min(Number.MIN_SAFE_INTEGER)
    .max(Number.MAX_SAFE_INTEGER),
  max: Yup.number()
    .required('This field is required.')
    .min(Number.MIN_SAFE_INTEGER)
    .max(Number.MAX_SAFE_INTEGER)
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { min } = this.parent;
        return !min || !value || value >= min;
      }
    ),
  precision: bigIntValidator
    .required('This field is required.')
    .test(
      'min',
      'Value must be greater than or equal to 1',
      getBigIntMinValidator(1)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${17}`,
      getBigIntMaxValidator(17)
    ),
});

const generateGenderConfig = Yup.object().shape({
  abbreviate: Yup.boolean().default(false).required('This field is required.'),
});

const generateInt64Config = Yup.object().shape({
  randomizeSign: Yup.bool().default(false).required('This field is required.'),
  min: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 1',
      getBigIntMinValidator(1)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { max } = this.parent;
      const maxBig = tryBigInt(max);
      const valueBig = tryBigInt(value ?? 0);
      return maxBig !== null && valueBig !== null && valueBig <= maxBig;
    })
    .required('This field is required.'),
  max: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 1',
      getBigIntMinValidator(1)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { min } = this.parent;
        const valueBig = tryBigInt(value ?? 0);
        const minBig = tryBigInt(min);
        return minBig !== null && valueBig !== null && valueBig >= minBig;
      }
    )
    .required('This field is required.'),
});

const generateStringPhoneNumberConfig = Yup.object().shape({
  min: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 8',
      getBigIntMinValidator(8)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { max } = this.parent;
      const maxBig = tryBigInt(max);
      const valueBig = tryBigInt(value ?? 0);
      return maxBig !== null && valueBig !== null && valueBig <= maxBig;
    })
    .required('This field is required.'),
  max: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 12',
      getBigIntMinValidator(12)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { min } = this.parent;
        const valueBig = tryBigInt(value ?? 0);
        const minBig = tryBigInt(min);
        return minBig !== null && valueBig !== null && valueBig >= minBig;
      }
    )
    .required('This field is required.'),
});

const generateStringConfig = Yup.object().shape({
  min: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 1',
      getBigIntMinValidator(1)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .required('This field is required.')
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { max } = this.parent;
      const maxBig = tryBigInt(max);
      const valueBig = tryBigInt(value);
      return maxBig !== null && valueBig !== null && valueBig <= maxBig;
    }),
  max: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 1',
      getBigIntMinValidator(1)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .required('This field is required.')
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { min } = this.parent;
        const valueBig = tryBigInt(value);
        const minBig = tryBigInt(min);
        return minBig !== null && valueBig !== null && valueBig >= minBig;
      }
    ),
});

const generateUuidConfig = Yup.object().shape({
  includeHyphens: Yup.boolean()
    .default(false)
    .required('This field is required.'),
});

const transformE164PhoneNumber = Yup.object().shape({
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
});

const transformFirstNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
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
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
});

const transformInt64PhoneNumberConfig = Yup.object().shape({
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
});

const transformInt64Config = Yup.object().shape({
  randomizationRangeMin: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 1',
      getBigIntMinValidator(1)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .required('This field is required.')
    .test('is-less-than-max', 'Min must be less than Max', function (value) {
      const { randomizationRangeMax } = this.parent;
      const maxBig = tryBigInt(randomizationRangeMax);
      const valueBig = tryBigInt(value);
      return maxBig !== null && valueBig !== null && valueBig <= maxBig;
    }),
  randomizationRangeMax: bigIntValidator
    .test(
      'min',
      'Value must be greater than or equal to 1',
      getBigIntMinValidator(1)
    )
    .test(
      'max',
      `Value must be less than than or equal to ${Number.MAX_SAFE_INTEGER}`,
      getBigIntMaxValidator(Number.MAX_SAFE_INTEGER)
    )
    .required('This field is required.')
    .test(
      'is-greater-than-min',
      'Max must be greater than Min',
      function (value) {
        const { randomizationRangeMin } = this.parent;
        const valueBig = tryBigInt(value);
        const minBig = tryBigInt(randomizationRangeMin);
        return minBig !== null && valueBig !== null && valueBig >= minBig;
      }
    ),
});

const transformLastNameConfig = Yup.object().shape({
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
});

const transformStringPhoneNumberConfig = Yup.object().shape({
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
});

const transformStringConfig = Yup.object().shape({
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
});

const userDefinedTransformerConfig = Yup.object().shape({
  id: Yup.string().required('This field is required.'),
});

const generateStateConfig = Yup.object().shape({
  generateFullName: Yup.boolean()
    .default(false)
    .required('This field is required'),
});

const JavascriptConfig = Yup.object().shape({
  code: Yup.string()
    .required('This field is required.')
    .test(
      'is-valid-javascript',
      'The JavaScript code is invalid.',
      async (value, context) => {
        const accountId = context?.options?.context?.accountId;
        if (!accountId) {
          return context.createError({
            message: 'Unable to verify Account Id.',
          });
        }
        try {
          const isUserJavascriptCodeValid:
            | UseMutateAsyncFunction<
                ValidateUserJavascriptCodeResponse,
                ConnectError,
                PartialMessage<ValidateUserJavascriptCodeRequest>,
                unknown
              >
            | undefined = context?.options?.context?.isUserJavascriptCodeValid;
          if (isUserJavascriptCodeValid) {
            const res = await isUserJavascriptCodeValid({
              accountId,
              code: value,
            });
            if (!res.valid) {
              return context.createError({
                message: 'Javascript is not valid',
              });
            }
          }
          return true;
        } catch (error) {
          return context.createError({
            message: `Unable to verify Javascript code: ${error}`,
          });
        }
      }
    ),
});

const generateCategoricalConfig = Yup.object().shape({
  categories: Yup.string().required('This field is required.'),
});

type ConfigType = TransformerConfig['config'];

// Helper function to extract the 'case' property from a config type
type ExtractCase<T> = T extends { case: infer U } ? U : never;

// Computed type that extracts all case types from the config union
type TransformerConfigCase = ExtractCase<ConfigType>;

const transformCharacterScrambleConfig = Yup.object().shape({
  userProvidedRegex: Yup.string().optional(),
});

// This is intended to be empty and is used for any transformer config that has no configuration options
const EMPTY_TRANSFORMER_VALUE_CONFIG = Yup.object({});

// Using this "as const" allows typescript to infer the types based on the shape we've described in the Yup object
// Ideally we can more explicitly type this in the future based on the Transformer types we get from @neosync/sdk
export const TRANSFORMER_SCHEMA_CONFIGS = {
  generateBoolConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateCityConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateDefaultConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateFirstNameConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateFullAddressConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateFullNameConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateInt64PhoneNumberConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateLastNameConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateSha256hashConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateSsnConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateStreetAddressConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateUnixtimestampConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateUsernameConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateUtctimestampConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateZipcodeConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  nullconfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  passthroughConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,

  generateEmailConfig: generateEmailConfig,
  transformEmailConfig: transformEmailConfig,
  generateCardNumberConfig: generateCardNumberConfig,
  generateE164PhoneNumberConfig: generateInternationalPhoneNumberConfig,
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
  transformPhoneNumberConfig: transformStringPhoneNumberConfig,
  transformStringConfig: transformStringConfig,
  userDefinedTransformerConfig: userDefinedTransformerConfig,
  transformJavascriptConfig: JavascriptConfig,
  generateCategoricalConfig: generateCategoricalConfig,
  transformCharacterScrambleConfig: transformCharacterScrambleConfig,
  generateJavascriptConfig: JavascriptConfig,
  generateStateConfig: generateStateConfig,
} as const;

// This is here so that whenever we add a new transformer, it errors due to the typing of the key to the TransformerConfigCase
const KEYED_TRANSFORMER_SCHEMA_CONFIGS: Record<
  NonNullable<TransformerConfigCase>,
  Yup.ObjectSchema<any> // eslint-disable-line @typescript-eslint/no-explicit-any
> = TRANSFORMER_SCHEMA_CONFIGS;

export const TransformerConfigSchema = Yup.lazy((v) => {
  const ccase = v?.case as TransformerConfigCase;
  if (!ccase) {
    return Yup.object({
      case: Yup.string().required(
        'A valid transformer configuration must be provided.'
      ),
      value: Yup.object().required(),
    });
  }
  const cconfig = KEYED_TRANSFORMER_SCHEMA_CONFIGS[ccase];
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
  .min(3)
  .max(30)
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

      if (!RESOURCE_NAME_REGEX.test(value)) {
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

      if (value === context?.options?.context?.name) {
        return true;
      }

      try {
        const isTransformerNameAvailable:
          | UseMutateAsyncFunction<
              IsTransformerNameAvailableResponse,
              ConnectError,
              PartialMessage<IsTransformerNameAvailableRequest>,
              unknown
            >
          | undefined = context?.options?.context?.isTransformerNameAvailable;
        if (isTransformerNameAvailable) {
          const res = await isTransformerNameAvailable({
            accountId,
            transformerName: value,
          });
          if (!res.isAvailable) {
            return context.createError({
              message: 'This Transformer Name is already taken.',
            });
          }
        }
        return true;
      } catch (error) {
        return context.createError({
          message: 'Error validating name availability.',
        });
      }
    }
  );

export const CreateUserDefinedTransformerFormValues = Yup.object({
  name: transformerNameSchema,
  source: Yup.number(),
  description: Yup.string().required(),
  config: TransformerConfigSchema,
});

export type CreateUserDefinedTransformerFormValues = Yup.InferType<
  typeof CreateUserDefinedTransformerFormValues
>;

export interface CreateUserDefinedTransformerFormContext {
  accountId: string;
  isTransformerNameAvailable: UseMutateAsyncFunction<
    IsTransformerNameAvailableResponse,
    ConnectError,
    PartialMessage<IsTransformerNameAvailableRequest>,
    unknown
  >;
  isJavascriptCodeValid: UseMutateAsyncFunction<
    ValidateUserJavascriptCodeResponse,
    ConnectError,
    PartialMessage<ValidateUserJavascriptCodeRequest>,
    unknown
  >;
}

export const UpdateUserDefinedTransformerFormValues = Yup.object({
  name: transformerNameSchema,
  id: Yup.string(),
  description: Yup.string().required(),
  config: TransformerConfigSchema,
});

export type UpdateUserDefinedTransformerFormValues = Yup.InferType<
  typeof UpdateUserDefinedTransformerFormValues
>;

export interface EditUserDefinedTransformerFormContext
  extends CreateUserDefinedTransformerFormContext {
  name: string;
}
