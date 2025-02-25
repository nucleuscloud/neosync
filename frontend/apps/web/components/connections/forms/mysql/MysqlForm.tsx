import { buildConnectionConfigMysql } from '@/app/(mgmt)/[account]/connections/util';
import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { MysqlFormValues } from '@/yup-validations/connections';
import { create as createMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  CheckConnectionConfigByIdRequestSchema,
  CheckConnectionConfigRequestSchema,
  ConnectionService,
} from '@neosync/sdk';
import { ReactElement, useEffect } from 'react';
import { create } from 'zustand';
import {
  CheckConnectionButton,
  ClientTlsAccordion,
  Name,
  SqlConnectionOptions,
  SshTunnelAccordion,
} from '../SharedFormInputs';
import { useHandleSubmit } from '../useHandleSubmit';
import DatabaseCredentials from './DatabaseCredentials';

interface MysqlFormStore extends BaseStore<MysqlFormValues> {
  init?(values: MysqlFormValues): void;
}

function getInitialFormState(): MysqlFormValues {
  return {
    activeTab: 'url',
    connectionName: '',
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
  connectionId?: string;
}

export default function MysqlForm(props: Props): ReactElement<any> {
  const {
    mode,
    initialValues,
    onSubmit = async () => undefined,
    canViewSecrets = false,
    getValueWithSecrets,
    connectionId,
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
    return MysqlFormValues.validate(values, {
      abortEarly: false,
      context: {
        accountId: account?.id ?? '',
        isConnectionNameAvailable: isConnectionNameAvailableAsync,
        originalConnectionName:
          mode === 'edit' ? initialValues?.connectionName : undefined,
        activeTab: values.activeTab ?? 'url',
      },
    });
  });

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
              connectionConfig: buildConnectionConfigMysql({
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
          connectionType="mysql"
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
