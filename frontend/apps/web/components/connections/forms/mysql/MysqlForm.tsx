import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { MysqlFormValues } from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import { FormEvent, ReactElement, useEffect } from 'react';
import { ValidationError } from 'yup';
import { create } from 'zustand';
import {
  ClientTlsAccordion,
  Name,
  SqlConnectionOptions,
  SshTunnelAccordion,
} from '../SharedFormInputs';
import DatabaseCredentials from './DatabaseCredentials';

interface MysqlFormStore extends BaseStore<MysqlFormValues> {
  init?(values: MysqlFormValues): void;
}

function getInitialFormState(): MysqlFormValues {
  return {
    connectionName: 'my-connection',
    url: '',
    envVar: '',
    db: {
      host: '',
      name: '',
      user: '',
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

const useFormStore = create<MysqlFormStore>((set) => ({
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
  initialValues?: MysqlFormValues;
  onSubmit?(values: MysqlFormValues): Promise<void>;
  canViewSecrets?: boolean;
  getValueWithSecrets?(): Promise<MysqlFormValues | undefined>;
}

export default function MysqlForm(props: Props): ReactElement {
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

      const validatedData = await MysqlFormValues.validate(formData, {
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

      {!isViewMode && <Submit isSubmitting={isSubmitting} text={submitText} />}
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
