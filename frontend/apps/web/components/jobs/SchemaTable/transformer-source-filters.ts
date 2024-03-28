type TransformerSource =
  | 'generate_email'
  | 'transform_email'
  | 'generate_bool'
  | 'generate_card_number'
  | 'generate_city'
  | 'generate_default'
  | 'generate_e164_phone_number'
  | 'generate_first_name'
  | 'generate_float64'
  | 'generate_full_address'
  | 'generate_full_name'
  | 'generate_gender'
  | 'generate_int64_phone_number'
  | 'generate_int64'
  | 'generate_last_name'
  | 'generate_sha256hash'
  | 'generate_ssn'
  | 'generate_state'
  | 'generate_street_address'
  | 'generate_string_phone_number'
  | 'generate_string'
  | 'generate_unixtimestamp'
  | 'generate_username'
  | 'generate_utctimestamp'
  | 'generate_uuid'
  | 'generate_zipcode'
  | 'transform_e164_phone_number'
  | 'transform_first_name'
  | 'transform_float64'
  | 'transform_full_name'
  | 'transform_int64_phone_number'
  | 'transform_int64'
  | 'transform_last_name'
  | 'transform_phone_number'
  | 'transform_string'
  | 'passthrough'
  | 'null'
  | 'transform_javascript'
  | 'generate_categorical'
  | 'transform_character_scramble';

interface FilterConfig {
  handlesNull: boolean;
  supportedDataTypes: Set<TransformerDataType>;
  supportedJobTypes: Set<JobType>;
}

type JobType = 'sync' | 'generate';
type TransformerDataType =
  | 'string'
  | 'boolean'
  | 'int64'
  | 'any'
  | 'float64'
  | 'time'
  | 'uuid'
  | 'null';

const TRANSFORMER_FILTER_CONFIG: Record<TransformerSource, FilterConfig> = {
  generate_bool: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_card_number: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_categorical: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_city: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_default: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_e164_phone_number: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_email: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_first_name: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_float64: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_full_address: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_full_name: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_gender: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_int64: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_int64_phone_number: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_last_name: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_sha256hash: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_ssn: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_state: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_street_address: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_string: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_string_phone_number: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_unixtimestamp: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_username: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_utctimestamp: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_uuid: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  generate_zipcode: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  null: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  passthrough: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_character_scramble: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_e164_phone_number: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_email: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_first_name: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_float64: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_full_name: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_int64: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_int64_phone_number: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_javascript: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_last_name: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_phone_number: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
  transform_string: {
    handlesNull: true,
    supportedJobTypes: new Set<JobType>(),
    supportedDataTypes: new Set<TransformerDataType>(),
  },
};
