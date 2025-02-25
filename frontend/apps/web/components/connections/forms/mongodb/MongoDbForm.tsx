import { buildConnectionConfigMongo } from '@/app/(mgmt)/[account]/connections/util';
import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { MongoDbFormValues } from '@/yup-validations/connections';
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
} from '../SharedFormInputs';
import { useHandleSubmit } from '../useHandleSubmit';
import DatabaseCredentials from './DatabaseCredentials';

interface MongoDbFormStore extends BaseStore<MongoDbFormValues> {
  init?(values: MongoDbFormValues): void;
}

function getInitialFormState(): MongoDbFormValues {
  return {
    connectionName: '',
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
  connectionId?: string;
}

export default function MongoDbForm(props: Props): ReactElement {
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
    return MongoDbFormValues.validate(values, {
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
