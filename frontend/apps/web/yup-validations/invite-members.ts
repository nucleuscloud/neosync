import * as Yup from 'yup';

export const InviteMembersForm = Yup.object({
  email: Yup.string().email().required('The Email is required'),
  role: Yup.number().required('The Role is required'),
});

export type InviteMembersForm = Yup.InferType<typeof InviteMembersForm>;
