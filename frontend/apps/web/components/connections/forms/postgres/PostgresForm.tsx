import { buildConnectionConfigPostgres } from '@/app/(mgmt)/[account]/connections/util';
import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { PostgresFormValues } from '@/yup-validations/connections';
import { create as createMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  CheckConnectionConfigByIdRequestSchema,
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
  SqlConnectionOptions,
  SshTunnelAccordion,
} from '../SharedFormInputs';
import DatabaseCredentials from './DatabaseCredentials';

interface PostgresFormStore extends BaseStore<PostgresFormValues> {
  init?(values: PostgresFormValues): void;
}

function getInitialFormState(): PostgresFormValues {
  return {
    connectionName: 'my-connection',
    url: '',
    envVar: '',
    db: {
      host: '',
      name: '',
      user: '',
      pass: '',
      port: 5432,
      sslMode: 'disable',
    },
    tunnel: {
      host: '',
      port: 22,
      knownHostPublicKey: '',
      user: '',
      passphrase: '',
      privateKey: '',
    },
    options: {
      maxConnectionLimit: 10,
      maxIdleDuration: '',
      maxIdleLimit: 2,
      maxOpenDuration: '',
    },
    clientTls: {
      clientCert: '',
      clientKey: '',
      rootCert: '',
    },
  };
}

const useFormStore = create<PostgresFormStore>((set) => ({
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
  initialValues?: PostgresFormValues;
  onSubmit?(values: PostgresFormValues): Promise<void>;
  canViewSecrets?: boolean;
  getValueWithSecrets?(): Promise<PostgresFormValues | undefined>;
  connectionId?: string;
}

export default function PostgresForm(props: Props): ReactElement {
  const {
    mode,
    initialValues,
    onSubmit,
    canViewSecrets = false,
    getValueWithSecrets,
    connectionId,
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

      const validatedData = await PostgresFormValues.validate(formData, {
        abortEarly: false,
        context: {
          accountId: account?.id ?? '',
          isConnectionNameAvailable: isConnectionNameAvailableAsync,
          originalConnectionName:
            mode === 'edit' ? initialValues?.connectionName : undefined,
          activeTab: formData.activeTab ?? 'url',
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
        activeTab={formData.activeTab ?? 'url'}
        setActiveTab={(value) => setFormData({ activeTab: value })}
        dbValue={formData.db}
        onDbValueChange={(value) =>
          isViewMode ? () => {} : setFormData({ db: value })
        }
        urlValue={formData.url}
        onUrlValueChange={(value) =>
          isViewMode ? () => {} : setFormData({ url: value })
        }
        envVarValue={formData.envVar}
        onEnvVarValueChange={(value) =>
          isViewMode ? () => {} : setFormData({ envVar: value })
        }
        isViewMode={isViewMode}
        canViewSecrets={canViewSecrets ?? false}
        onRevealClick={async () => getValueWithSecrets?.()}
      />
      <SqlConnectionOptions
        value={formData.options}
        onChange={(value) =>
          isViewMode ? () => {} : setFormData({ options: value })
        }
        errors={errors}
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
      <SshTunnelAccordion
        value={formData.tunnel}
        onChange={(value) =>
          isViewMode ? () => {} : setFormData({ tunnel: value })
        }
        errors={errors}
        isViewMode={isViewMode}
        canViewSecrets={canViewSecrets ?? false}
        onRevealClick={async () => {
          const values = await getValueWithSecrets?.();
          return values?.tunnel;
        }}
      />

      <div className="flex justify-end gap-3">
        <CheckConnectionButton
          isValid={Object.keys(errors).length === 0}
          getRequest={() => {
            return createMessage(CheckConnectionConfigRequestSchema, {
              connectionConfig: buildConnectionConfigPostgres({
                ...formData,
                url: formData.activeTab === 'url' ? formData.url : undefined,
                db: formData.db,
                envVar:
                  formData.activeTab === 'url-env'
                    ? formData.envVar
                    : undefined,
              }),
            });
          }}
          connectionName={formData.connectionName}
          connectionType="postgres"
          mode={mode === 'view' ? 'checkById' : 'check'}
          getRequestById={() => {
            return createMessage(CheckConnectionConfigByIdRequestSchema, {
              id: connectionId ?? '',
            });
          }}
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
