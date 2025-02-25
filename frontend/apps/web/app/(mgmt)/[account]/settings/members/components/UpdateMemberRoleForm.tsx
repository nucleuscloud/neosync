import ButtonText from '@/components/ButtonText';
import FormErrorMessage from '@/components/FormErrorMessage';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { BaseHookStore } from '@/util/zustand.stores.util';
import { UpdateMemberRoleFormValues } from '@/yup-validations/invite-members';
import { AccountRole, AccountUser } from '@neosync/sdk';
import { FormEvent, ReactElement, useEffect } from 'react';
import * as yup from 'yup';
import { create } from 'zustand';
import FormHeader from '../../../jobs/[id]/hooks/components/FormHeader';
import SelectAccountRole from './SelectAccountRole';

interface UpdateMemberRoleFormStore
  extends BaseHookStore<UpdateMemberRoleFormValues> {}

function getInitialFormState(): UpdateMemberRoleFormValues {
  return {
    role: AccountRole.JOB_VIEWER,
  };
}

const useStore = create<UpdateMemberRoleFormStore>((set) => ({
  formData: getInitialFormState(),
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () =>
    set({
      formData: getInitialFormState(),
      errors: {},
      isSubmitting: false,
    }),
}));

function toFormData(
  member: Pick<AccountUser, 'role'>
): UpdateMemberRoleFormValues {
  return {
    role: member.role,
  };
}

interface Props {
  member: Pick<AccountUser, 'role'>;
  onSubmit(values: UpdateMemberRoleFormValues): Promise<void>;
  onCancel(): void;
}

export default function UpdateMemberRoleForm(props: Props): ReactElement {
  const { member, onSubmit, onCancel } = props;
  const {
    formData,
    setFormData,
    errors,
    setErrors,
    isSubmitting,
    setSubmitting,
  } = useStore();

  useEffect(() => {
    // Initialize form with hook data
    const formData = toFormData(member);
    setFormData(formData);
  }, [member, setFormData]);

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    if (isSubmitting) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await UpdateMemberRoleFormValues.validate(
        formData,
        {
          abortEarly: false,
        }
      );

      await onSubmit(validatedData);
    } catch (err) {
      if (err instanceof yup.ValidationError) {
        const validationErrors: Record<string, string> = {};
        err.inner.forEach((error) => {
          if (error.path) {
            validationErrors[error.path] = error.message;
          }
        });
        setErrors(validationErrors);
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <AccountRoleField
        error={errors.role}
        value={formData.role}
        onChange={(role) => setFormData({ role })}
      />

      <div className="flex justify-end gap-3">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          className="w-full sm:w-auto"
        >
          <ButtonText text="Cancel" />
        </Button>

        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full sm:w-auto"
        >
          <ButtonText
            leftIcon={isSubmitting ? <Spinner /> : undefined}
            text="Update"
          />
        </Button>
      </div>
    </form>
  );
}

interface AccountRoleFieldProps {
  error?: string;
  value: AccountRole;
  onChange(value: AccountRole): void;
}
function AccountRoleField(props: AccountRoleFieldProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="role"
        title="Role"
        description="The new role of the user"
        isErrored={!!error}
      />
      <SelectAccountRole role={value} onChange={(role) => onChange(role)} />
      <FormErrorMessage message={error} />
    </div>
  );
}
