import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import SystemLicenseAlert from '@/components/SystemLicenseAlert';
import { BaseStore } from '@/util/zustand.stores.util';
import { GcpCloudStorageFormValues } from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import { FormEvent, ReactElement, useEffect } from 'react';
import { ValidationError } from 'yup';
import { create } from 'zustand';
import { Name } from '../SharedFormInputs';
import Bucket from './Bucket';

interface GcpCloudStorageFormStore
  extends BaseStore<GcpCloudStorageFormValues> {
  init?(values: GcpCloudStorageFormValues): void;
}

function getInitialFormState(): GcpCloudStorageFormValues {
  return {
    connectionName: 'my-connection',
    gcp: {
      bucket: '',
      pathPrefix: '',
    },
  };
}

const useFormStore = create<GcpCloudStorageFormStore>((set) => ({
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
  init: (values) => set({ formData: values }),
}));

type Mode = 'create' | 'edit' | 'view';

interface Props {
  mode: Mode;
  initialValues?: GcpCloudStorageFormValues;
  onSubmit?(values: GcpCloudStorageFormValues): Promise<void>;
  canViewSecrets?: boolean;
  getValueWithSecrets?(): Promise<GcpCloudStorageFormValues | undefined>;
}

export default function GcpCloudStorageForm(props: Props): ReactElement {
  const { mode, initialValues, onSubmit } = props;
  const { account } = useAccount();
  const {
    formData,
    errors,
    setFormData,
    setErrors,
    setSubmitting,
    isSubmitting,
    init,
    resetForm,
  } = useFormStore();

  useEffect(() => {
    if (initialValues) {
      init?.(initialValues);
    } else {
      resetForm();
    }
  }, [initialValues, init, resetForm]);

  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    ConnectionService.method.isConnectionNameAvailable
  );

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    if (isSubmitting || !onSubmit) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await GcpCloudStorageFormValues.validate(formData, {
        abortEarly: false,
        context: {
          accountId: account?.id ?? '',
          isConnectionNameAvailable: isConnectionNameAvailableAsync,
          originalConnectionName:
            mode === 'edit' ? initialValues?.connectionName : undefined,
        },
      });

      await onSubmit(validatedData);
    } catch (err) {
      if (err instanceof ValidationError) {
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

  const isViewMode = mode === 'view';
  const submitText = mode === 'create' ? 'Create' : 'Update';

  const formContent = (
    <>
      <SystemLicenseAlert />

      <Name
        error={errors.connectionName}
        value={formData.connectionName}
        onChange={
          isViewMode
            ? () => {}
            : (value) => setFormData({ connectionName: value })
        }
      />

      <Bucket
        value={formData.gcp}
        onChange={(value) => setFormData({ gcp: value })}
        errors={errors}
      />

      <div className="flex justify-end gap-3">
        {!isViewMode && (
          <Submit isSubmitting={isSubmitting} text={submitText} />
        )}
      </div>
    </>
  );

  if (isViewMode) {
    return <div className="space-y-6">{formContent}</div>;
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {formContent}
    </form>
  );
}
