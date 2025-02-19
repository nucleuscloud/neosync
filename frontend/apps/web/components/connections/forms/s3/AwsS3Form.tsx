import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { AwsFormValues } from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import { FormEvent, ReactElement, useEffect } from 'react';
import { ValidationError } from 'yup';
import { create } from 'zustand';
import {
  AwsAdvancedConfigAccordion,
  AwsCredentialsFormAccordion,
  Name,
} from '../SharedFormInputs';
import Bucket from './Bucket';

interface AwsS3FormStore extends BaseStore<AwsFormValues> {
  init?(values: AwsFormValues): void;
}

function getInitialFormState(): AwsFormValues {
  return {
    connectionName: 'my-connection',
    s3: {
      bucket: '',
      pathPrefix: '',
    },
    advanced: {
      region: '',
      endpoint: '',
    },
    credentials: {
      accessKeyId: '',
      secretAccessKey: '',
      sessionToken: '',
      fromEc2Role: false,
      roleArn: '',
      roleExternalId: '',
      profile: '',
    },
  };
}

const useFormStore = create<AwsS3FormStore>((set) => ({
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
  initialValues?: AwsFormValues;
  onSubmit?(values: AwsFormValues): Promise<void>;
  canViewSecrets?: boolean;
  getValueWithSecrets?(): Promise<AwsFormValues | undefined>;
}

export default function AwsS3Form(props: Props): ReactElement {
  const {
    mode,
    initialValues,
    onSubmit,
    canViewSecrets = false,
    getValueWithSecrets,
  } = props;
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

      const validatedData = await AwsFormValues.validate(formData, {
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
        value={formData.s3}
        onChange={(value) => setFormData({ s3: value })}
        errors={errors}
      />

      <AwsAdvancedConfigAccordion
        value={formData.advanced ?? {}}
        onChange={(value) => setFormData({ advanced: value })}
        errors={errors}
      />

      <AwsCredentialsFormAccordion
        value={formData.credentials ?? {}}
        onChange={(value) => setFormData({ credentials: value })}
        isViewMode={isViewMode}
        canViewSecrets={canViewSecrets}
        onRevealClick={async () => {
          const values = await getValueWithSecrets?.();
          return values?.credentials ?? {};
        }}
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
