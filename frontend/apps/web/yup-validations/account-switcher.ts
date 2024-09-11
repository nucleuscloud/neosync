import * as Yup from 'yup';
import { RESOURCE_NAME_REGEX } from './connections';

export const CreateTeamFormValues = Yup.object({
  name: Yup.string()
    .required('The Name is required.')
    .min(3, 'The Name must be at least 3 characters.')
    .max(30, 'The Name cannot be greater than 30 characters.')
    .test(
      'valid account name',
      'Account Name must be of length 3-30 and only include lowercased letters, numbers, and/or hyphens.',
      (value) => {
        if (!value || value.length < 3) {
          return false;
        }
        if (!RESOURCE_NAME_REGEX.test(value)) {
          return false;
        }
        // todo: test to make sure that account is valid across neosync
        return true;
      }
    ),
});
export type CreateTeamFormValues = Yup.InferType<typeof CreateTeamFormValues>;
