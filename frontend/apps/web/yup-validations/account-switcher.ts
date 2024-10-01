import * as Yup from 'yup';
import { RESOURCE_NAME_REGEX } from './connections';

export const CreateTeamFormValues = Yup.object({
  name: Yup.string()
    .required('The Name is required.')
    .min(3, 'The Name must be at least 3 characters.')
    .max(100, 'The Name must be less than or equal to 100 characters.')
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
  convertPersonalToTeam: Yup.boolean().default(false),
});
export type CreateTeamFormValues = Yup.InferType<typeof CreateTeamFormValues>;
