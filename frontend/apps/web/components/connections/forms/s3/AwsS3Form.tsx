import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import SystemLicenseAlert from '@/components/SystemLicenseAlert';
import { BaseStore } from '@/util/zustand.stores.util';
import { AwsFormValues } from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import { ReactElement, useEffect } from 'react';
import { create } from 'zustand';
import {
  AwsAdvancedConfigAccordion,
  AwsCredentialsFormAccordion,
  Name,
} from '../SharedFormInputs';
import { useHandleSubmit } from '../useHandleSubmit';
import Bucket from './Bucket';

interface AwsS3FormStore extends BaseStore<AwsFormValues> {
  init?(values: AwsFormValues): void;
}

function getInitialFormState(): AwsFormValues {
  return {
    connectionName: '',
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

export default function AwsS3Form(props: Props): ReactElement<any> {
  const {
    mode,
    initialValues,
    onSubmit = async () => undefined,
    canViewSecrets = false,
    getValueWithSecrets,
  } = props;
  const { account } = useAccount();
  const store = useFormStore();

  const { formData, errors, isSubmitting, setFormData, resetForm, init } =
    store;

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

  const handleSubmit = useHandleSubmit(store, onSubmit, async (values) => {
    return AwsFormValues.validate(values, {
      abortEarly: false,
      context: {
        accountId: account?.id ?? '',
        isConnectionNameAvailable: isConnectionNameAvailableAsync,
        originalConnectionName:
          mode === 'edit' ? initialValues?.connectionName : undefined,
      },
    });
  });

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
