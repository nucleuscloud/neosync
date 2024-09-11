import * as Yup from 'yup';
import { RESOURCE_NAME_REGEX } from './connections';

export const ApiKeyFormValues = Yup.object({
  name: Yup.string()
    .required('Name is a required field.')
    .min(3, 'Name cannot be less than 3 characters.')
    .max(100, 'Name must be less than or equal to 100 characters.')
    .test(
      'validApiKeyName',
      'API Key Name must be at least 3 characters long and can only include lowercase letters, numbers, and hyphens.',
      (value) => {
        if (!value || value.length < 3) {
          return false;
        }
        if (!RESOURCE_NAME_REGEX.test(value)) {
          return false;
        }
        // todo: add server-side check to see if it's available on the backend
        return true;
      }
    ),
  expiresAtSelect: Yup.string().oneOf(['7', '30', '60', '90', 'custom']),
  expiresAt: Yup.date().required('The Expiration is a required field.'),
});

export type ApiKeyFormValues = Yup.InferType<typeof ApiKeyFormValues>;

export const RegenerateApiKeyForm = Yup.object({
  expiresAtSelect: Yup.string().oneOf(['7', '30', '60', '90', 'custom']),
  expiresAt: Yup.date().required('The Expiration is required.'),
});
export type RegenerateApiKeyForm = Yup.InferType<typeof RegenerateApiKeyForm>;
