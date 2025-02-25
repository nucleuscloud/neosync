import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { OpenAiFormValues } from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import { ReactElement, useEffect } from 'react';
import { create } from 'zustand';
import { Name } from '../SharedFormInputs';
import { useHandleSubmit } from '../useHandleSubmit';
import Sdk from './Sdk';

interface OpenAiFormStore extends BaseStore<OpenAiFormValues> {
  init?(values: OpenAiFormValues): void;
}

function getInitialFormState(): OpenAiFormValues {
  return {
    connectionName: '',
    sdk: {
      url: 'https://api.openai.com/v1',
      apiKey: '',
    },
  };
}

const useFormStore = create<OpenAiFormStore>((set) => ({
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
  init: (values: OpenAiFormValues) => set({ formData: values }),
}));

type Mode = 'create' | 'edit' | 'view';

interface Props {
  mode: Mode;
  initialValues?: OpenAiFormValues;
  onSubmit?(values: OpenAiFormValues): Promise<void>;
  canViewSecrets?: boolean;
  getValueWithSecrets?(): Promise<OpenAiFormValues | undefined>;
}

export default function OpenAiForm(props: Props): ReactElement<any> {
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
    return OpenAiFormValues.validate(values, {
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

  async function onRevealPassword(): Promise<string> {
    const values = await getValueWithSecrets?.();
    return values?.sdk.apiKey ?? '';
  }

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
      <Sdk
        errors={errors}
        value={formData.sdk}
        onChange={
          isViewMode ? () => {} : (value) => setFormData({ sdk: value })
        }
        isViewMode={isViewMode}
        canViewSecrets={canViewSecrets}
        onRevealPassword={onRevealPassword}
      />

      {!isViewMode && (
        <div className="flex justify-end gap-3">
          <Submit isSubmitting={isSubmitting} text={submitText} />
        </div>
      )}
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
