import { buildConnectionConfigOpenAi } from '@/app/(mgmt)/[account]/connections/util';
import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { OpenAiFormValues } from '@/yup-validations/connections';
import { create as createMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  ConnectionService,
  CreateConnectionRequest,
  CreateConnectionRequestSchema,
} from '@neosync/sdk';
import { FormEvent, ReactElement } from 'react';
import { ValidationError } from 'yup';
import { create } from 'zustand';
import { Name } from '../SharedFormInputs';
import Sdk from './Sdk';

interface NewOpenAiConnectionStore extends BaseStore<OpenAiFormValues> {}

function getInitialNewFormState(): OpenAiFormValues {
  return {
    connectionName: 'my-connection',
    sdk: {
      url: 'https://api.openai.com/v1',
      apiKey: '',
    },
  };
}

const useNewHookStore = create<NewOpenAiConnectionStore>((set) => ({
  formData: getInitialNewFormState(),
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () =>
    set({
      formData: getInitialNewFormState(),
      errors: {},
      isSubmitting: false,
    }),
}));

interface Props {
  onSubmit(values: CreateConnectionRequest): Promise<void>;
}

export default function NewOpenAiForm(props: Props): ReactElement {
  const { onSubmit } = props;
  const { account } = useAccount();
  const {
    formData,
    errors,
    setFormData,
    setErrors,
    setSubmitting,
    isSubmitting,
  } = useNewHookStore();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    ConnectionService.method.isConnectionNameAvailable
  );

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    if (isSubmitting) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await OpenAiFormValues.validate(formData, {
        abortEarly: false,
        context: {
          accountId: account?.id ?? '',
          isConnectionNameAvailable: isConnectionNameAvailableAsync,
        },
      });
      await onSubmit(
        createMessage(CreateConnectionRequestSchema, {
          accountId: account?.id ?? '',
          name: validatedData.connectionName,
          connectionConfig: buildConnectionConfigOpenAi(validatedData),
        })
      );
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

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Name
        error={errors.connectionName}
        value={formData.connectionName}
        onChange={(value) => setFormData({ connectionName: value })}
      />
      <Sdk
        errors={errors}
        value={formData.sdk}
        onChange={(value) => setFormData({ sdk: value })}
      />

      <Submit isSubmitting={isSubmitting} text="Create" />
    </form>
  );
}
