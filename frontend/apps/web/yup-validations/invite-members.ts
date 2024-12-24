import * as Yup from 'yup';

const RoleFormValue = Yup.number().required('The Role is required');

export const InviteMembersForm = Yup.object({
  email: Yup.string().email().required('The Email is required'),
  role: RoleFormValue,
});

export type InviteMembersForm = Yup.InferType<typeof InviteMembersForm>;

export const UpdateMemberRoleFormValues = Yup.object({
  role: RoleFormValue,
});

export type UpdateMemberRoleFormValues = Yup.InferType<
  typeof UpdateMemberRoleFormValues
>;
