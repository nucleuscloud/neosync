import { buildConnectionConfigMongo } from '@/app/(mgmt)/[account]/connections/util';
import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { MongoDbFormValues } from '@/yup-validations/connections';
import { create as createMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  CheckConnectionConfigRequestSchema,
  ConnectionService,
} from '@neosync/sdk';
import { FormEvent, ReactElement, useEffect } from 'react';
import { ValidationError } from 'yup';
import { create } from 'zustand';
import {
  CheckConnectionButton,
  ClientTlsAccordion,
  Name,
} from '../SharedFormInputs';
import DatabaseCredentials from './DatabaseCredentials';

interface MongoDbFormStore extends BaseStore<MongoDbFormValues> {
  init?(values: MongoDbFormValues): void;
}

function getInitialFormState(): MongoDbFormValues {
  return {
    connectionName: 'my-connection',
    url: '',
    clientTls: {
      clientCert: '',
      clientKey: '',
      rootCert: '',
    },
  };
}

const useFormStore = create<MongoDbFormStore>((set) => ({
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
  initialValues?: MongoDbFormValues;
  onSubmit?(values: MongoDbFormValues): Promise<void>;
  canViewSecrets?: boolean;
  getValueWithSecrets?(): Promise<MongoDbFormValues | undefined>;
}

export default function MongoDbForm(props: Props): ReactElement {
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

      const validatedData = await MongoDbFormValues.validate(formData, {
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
      <DatabaseCredentials
        errors={errors}
        urlValue={formData.url}
        onUrlValueChange={(value) =>
          isViewMode ? () => {} : setFormData({ url: value })
        }
        isViewMode={isViewMode}
        canViewSecrets={canViewSecrets ?? false}
        onRevealClick={async () => getValueWithSecrets?.()}
      />
      <ClientTlsAccordion
        value={formData.clientTls}
        onChange={(value) =>
          isViewMode ? () => {} : setFormData({ clientTls: value })
        }
        errors={errors}
        isViewMode={isViewMode}
        canViewSecrets={canViewSecrets ?? false}
        onRevealClick={async () => {
          const values = await getValueWithSecrets?.();
          return values?.clientTls;
        }}
      />

      <div className="flex justify-end gap-3">
        <CheckConnectionButton
          isValid={Object.keys(errors).length === 0}
          getRequest={() => {
            return createMessage(CheckConnectionConfigRequestSchema, {
              connectionConfig: buildConnectionConfigMongo({
                ...formData,
              }),
            });
          }}
          connectionName={formData.connectionName}
          connectionType="mongodb"
        />
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
