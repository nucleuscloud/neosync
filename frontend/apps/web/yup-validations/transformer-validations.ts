import {
  getBigIntValidateMaxFn,
  getBigIntValidateMinFn,
  getBigIntValidator,
} from '@/yup-validations/bigint';
import { RESOURCE_NAME_REGEX } from '@/yup-validations/connections';
import {
  getNumberValidateMaxFn,
  getNumberValidateMinFn,
} from '@/yup-validations/number';
import { create, MessageInitShape } from '@bufbuild/protobuf';
import {
  ConnectError,
  IsTransformerNameAvailableRequestSchema,
  IsTransformerNameAvailableResponse,
  TransformerConfig,
  ValidateUserJavascriptCodeRequestSchema,
  ValidateUserJavascriptCodeResponse,
} from '@neosync/sdk';
import { UseMutateAsyncFunction } from '@tanstack/react-query';
import * as Yup from 'yup';

const transformEmailConfig = Yup.object().shape({
  preserveDomain: Yup.boolean()
    .default(false)
    .required('This field is required.'),
  preserveLength: Yup.boolean()
    .default(false)
    .required('This field is required.'),
  excludedDomains: Yup.array()
    .of(
      Yup.string().required(
        'A non-empty domain is required in the excluded domains property'
      )
    )
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

const generateInternationalPhoneNumberConfig = Yup.object({
  min: getBigIntValidator({
    range: [9, 15],
  }).test(
    'is-less-than-or-equal-to-max',
    'Min must be less than or equal to Max',
    function (value) {
      const { max } = this.parent;
      const maxValidator = getBigIntValidateMaxFn(max);
      return maxValidator(value);
    }
  ),
  max: getBigIntValidator({
    range: [9, 15],
  }).test(
    'is-greater-than-or-equal-to-min',
    'Max must be greater than or equal to Min',
    function (value) {
      const { min } = this.parent;
      const minValidator = getBigIntValidateMinFn(min);
      return minValidator(value);
    }
  ),
});

const generateFloat64Config = Yup.object().shape({
  randomizeSign: Yup.bool().default(false),
  min: Yup.number()
    .min(
      Number.MIN_SAFE_INTEGER,
      'The Minimum cannot be less than −9007199254740991 (−(2^53 − 1))'
    )
    .max(
      Number.MAX_SAFE_INTEGER,
      'The Maximum cannot be greater than 9007199254740991 2^53 − 1'
    )
    .test(
      'is-less-than-or-equal-to-max',
      'Min must be less than or equal to Max',
      function (value) {
        const { max } = this.parent;
        const maxValidator = getNumberValidateMaxFn(max);
        return maxValidator(value);
      }
    ),
  max: Yup.number()
    .min(
      Number.MIN_SAFE_INTEGER,
      'The Minimum cannot be less than −9007199254740991 (−(2^53 − 1))'
    )
    .max(
      Number.MAX_SAFE_INTEGER,
      'The Maximum cannot be greater than 9007199254740991 2^53 − 1'
    )
    .test(
      'is-greater-than-or-equal-to-min',
      'Max must be greater than or equal to Min',
      function (value) {
        const { min } = this.parent;
        const minValidator = getNumberValidateMinFn(min);
        return minValidator(value);
      }
    ),
  precision: getBigIntValidator({
    range: [1, 17],
  }),
});

const generateGenderConfig = Yup.object().shape({
  abbreviate: Yup.boolean().default(false).required('This field is required.'),
});

const generateInt64Config = Yup.object().shape({
  randomizeSign: Yup.bool().default(false).required('This field is required.'),
  min: getBigIntValidator({
    range: [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER],
  }).test(
    'is-less-than-or-equal-to-max',
    'Min must be less than or equal to Max',
    function (value) {
      const { max } = this.parent;
      const maxValidator = getBigIntValidateMaxFn(max);
      return maxValidator(value);
    }
  ),
  max: getBigIntValidator({
    range: [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER],
  }).test(
    'is-greater-than-or-equal-to-min',
    'Max must be greater than or equal to Min',
    function (value) {
      const { min } = this.parent;
      const minValidator = getBigIntValidateMinFn(min);
      return minValidator(value);
    }
  ),
});

const generateStringPhoneNumberConfig = Yup.object().shape({
  min: getBigIntValidator({
    range: [8, 14],
  }).test(
    'is-less-than-or-equal-to-max',
    'Min must be less than or equal to Max',
    function (value) {
      const { max } = this.parent;
      const maxValidator = getBigIntValidateMaxFn(max);
      return maxValidator(value);
    }
  ),
  max: getBigIntValidator({
    range: [8, 14],
  }).test(
    'is-greater-than-or-equal-to-min',
    'Max must be greater than or equal to Min',
    function (value) {
      const { min } = this.parent;
      const minValidator = getBigIntValidateMinFn(min);
      return minValidator(value);
    }
  ),
});

const generateStringConfig = Yup.object().shape({
  min: getBigIntValidator({
    default: 1,
    requiredMessage:
      'Must provide a minimum number for generate string config.',
    range: [1, Number.MAX_SAFE_INTEGER],
  }).test(
    'is-less-than-or-equal-to-max',
    'Min must be less than or equal to Max',
    function (value) {
      const { max } = this.parent;
      const maxValidator = getBigIntValidateMaxFn(max);
      return maxValidator(value);
    }
  ),
  max: getBigIntValidator({
    default: 100,
    requiredMessage:
      'Must provide a maximum number for generate string config.',
    range: [1, Number.MAX_SAFE_INTEGER],
  }).test(
    'is-greater-than-or-equal-to-min',
    'Max must be greater than or equal to Min',
    function (value) {
      const { min } = this.parent;
      const minValidator = getBigIntValidateMinFn(min);
      return minValidator(value);
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
  randomizationRangeMin: Yup.number().test(
    'is-less-than-or-equal-to-max',
    'Min must be less than or equal to Max',
    function (value) {
      const { randomizationRangeMax } = this.parent;
      const maxValidator = getNumberValidateMaxFn(randomizationRangeMax);
      return maxValidator(value);
    }
  ),
  randomizationRangeMax: Yup.number().test(
    'is-greater-than-or-equal-to-min',
    'Max must be greater than or equal to Min',
    function (value) {
      const { randomizationRangeMin } = this.parent;
      const minValidator = getNumberValidateMinFn(randomizationRangeMin);
      return minValidator(value);
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
  randomizationRangeMin: getBigIntValidator({
    range: [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER],
  }).test(
    'is-less-than-or-equal-to-max',
    'Min must be less than or equal to Max',
    function (value) {
      const { randomizationRangeMax } = this.parent;
      const maxValidator = getBigIntValidateMaxFn(randomizationRangeMax);
      return maxValidator(value);
    }
  ),
  randomizationRangeMax: getBigIntValidator({
    range: [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER],
  }).test(
    'is-greater-than-or-equal-to-min',
    'Max must be greater than or equal to Min',
    function (value) {
      const { randomizationRangeMin } = this.parent;
      const minValidator = getBigIntValidateMinFn(randomizationRangeMin);
      return minValidator(value);
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
        try {
          const isUserJavascriptCodeValid:
            | UseMutateAsyncFunction<
                ValidateUserJavascriptCodeResponse,
                ConnectError,
                MessageInitShape<
                  typeof ValidateUserJavascriptCodeRequestSchema
                >,
                unknown
              >
            | undefined = context?.options?.context?.isUserJavascriptCodeValid;
          if (isUserJavascriptCodeValid) {
            const res = await isUserJavascriptCodeValid(
              create(ValidateUserJavascriptCodeRequestSchema, {
                code: value,
              })
            );
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
  categories: Yup.string()
    .min(1, 'Must have at least one category')
    .required(
      'categories is a required field in the Generate Categorical config.'
    ),
});

const generateCountryConfig = Yup.object().shape({
  generateFullName: Yup.boolean()
    .default(false)
    .required('This field is required'),
});

const transformPiiTextConfig = Yup.object({
  scoreThreshold: Yup.number()
    .min(0, 'Must not go below 0')
    .max(1, 'Must not go above 1')
    .required('This field is required.'),
  // todo: add in default anonymizer fields
});

const generateIpAddressConfig = Yup.object({
  ipType: Yup.string().default('GENERATE_IP_ADDRESS_TYPE_V4_PUBLIC'),
});

const transfromHashConfig = Yup.object({
  algo: Yup.string().default('TRANSFORM_HASH_ALGO_SHA256'),
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
const TRANSFORMER_SCHEMA_CONFIGS = {
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
  generateCountryConfig: generateCountryConfig,
  generateBusinessNameConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  generateIpAddressConfig: generateIpAddressConfig,
  transformUuidConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  transformScrambleIdentityConfig: EMPTY_TRANSFORMER_VALUE_CONFIG,
  transformHashConfig: transfromHashConfig,

  transformPiiTextConfig: transformPiiTextConfig,
} as const;

// This is here so that whenever we add a new transformer, it errors due to the typing of the key to the TransformerConfigCase
const KEYED_TRANSFORMER_SCHEMA_CONFIGS: Record<
  NonNullable<TransformerConfigCase>,
  Yup.ObjectSchema<any> // eslint-disable-line @typescript-eslint/no-explicit-any
> = TRANSFORMER_SCHEMA_CONFIGS;

export const TransformerConfigFormValue = Yup.lazy((v) => {
  const ccase = v?.case as TransformerConfigCase;
  if (!ccase) {
    return Yup.object({
      case: Yup.string().required(
        'A valid transformer configuration must be provided.'
      ),
      value: Yup.object().required('The Transformer value is required.'),
    });
  }
  const cconfig = KEYED_TRANSFORMER_SCHEMA_CONFIGS[ccase];
  return Yup.object({
    case: Yup.string()
      .required('The Transformer case is required.')
      .oneOf([ccase]),
    value: cconfig,
  });
});

// Simplified version of a job mapping transformer config for use with react-hook-form only
export type TransformerConfigFormValue = Yup.InferType<
  typeof TransformerConfigFormValue
>;

const transformerNameSchema = Yup.string()
  .required('The Transformer Name is required.')
  .min(3, 'The Transformer Name must be at least 3 characters.')
  .max(
    100,
    'The Transformer Name must be less than or equal to 100 characters.'
  )
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
              MessageInitShape<typeof IsTransformerNameAvailableRequestSchema>,
              unknown
            >
          | undefined = context?.options?.context?.isTransformerNameAvailable;
        if (isTransformerNameAvailable) {
          const res = await isTransformerNameAvailable(
            create(IsTransformerNameAvailableRequestSchema, {
              accountId,
              transformerName: value,
            })
          );
          if (!res.isAvailable) {
            return context.createError({
              message: 'This Transformer Name is already taken.',
            });
          }
        }
        return true;
      } catch (error) {
        return context.createError({
          message: `Error validating name availability: ${error}`,
        });
      }
    }
  );

export const CreateUserDefinedTransformerFormValues = Yup.object({
  name: transformerNameSchema,
  source: Yup.number(),
  description: Yup.string().required('Description is a required field.'),
  config: TransformerConfigFormValue,
});

export type CreateUserDefinedTransformerFormValues = Yup.InferType<
  typeof CreateUserDefinedTransformerFormValues
>;

export interface CreateUserDefinedTransformerFormContext {
  accountId: string;
  isTransformerNameAvailable: UseMutateAsyncFunction<
    IsTransformerNameAvailableResponse,
    ConnectError,
    MessageInitShape<typeof IsTransformerNameAvailableRequestSchema>,
    unknown
  >;
  isUserJavascriptCodeValid: UseMutateAsyncFunction<
    ValidateUserJavascriptCodeResponse,
    ConnectError,
    MessageInitShape<typeof ValidateUserJavascriptCodeRequestSchema>,
    unknown
  >;
}

export const EditJobMappingTransformerConfigFormValues = Yup.object({
  config: TransformerConfigFormValue,
}).required('The Transformer config is required.');
export type EditJobMappingTransformerConfigFormValues = Yup.InferType<
  typeof EditJobMappingTransformerConfigFormValues
>;

export interface EditJobMappingTransformerConfigFormContext {
  accountId: string;
  isUserJavascriptCodeValid: UseMutateAsyncFunction<
    ValidateUserJavascriptCodeResponse,
    ConnectError,
    MessageInitShape<typeof ValidateUserJavascriptCodeRequestSchema>,
    unknown
  >;
}

export const UpdateUserDefinedTransformerFormValues = Yup.object({
  name: transformerNameSchema,
  id: Yup.string(),
  description: Yup.string().required('The Description is required.'),
  config: TransformerConfigFormValue,
});

export type UpdateUserDefinedTransformerFormValues = Yup.InferType<
  typeof UpdateUserDefinedTransformerFormValues
>;

export interface EditUserDefinedTransformerFormContext
  extends CreateUserDefinedTransformerFormContext {
  name: string;
}
