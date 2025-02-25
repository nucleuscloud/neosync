import { buildConnectionConfigDynamoDB } from '@/app/(mgmt)/[account]/connections/util';
import Submit from '@/components/forms/Submit';
import { useAccount } from '@/components/providers/account-provider';
import { BaseStore } from '@/util/zustand.stores.util';
import { DynamoDbFormValues } from '@/yup-validations/connections';
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
  AwsAdvancedConfigAccordion,
  AwsCredentialsFormAccordion,
  CheckConnectionButton,
  Name,
} from '../SharedFormInputs';
import { useHandleSubmit } from '../useHandleSubmit';

interface DynamoDbFormStore extends BaseStore<DynamoDbFormValues> {
  init?(values: DynamoDbFormValues): void;
}

function getInitialFormState(): DynamoDbFormValues {
  return {
    connectionName: '',
    advanced: {
      region: '',
      endpoint: '',
    },
    credentials: {
      accessKeyId: '',
      secretAccessKey: '',
      sessionToken: '',
    },
  };
}

const useFormStore = create<DynamoDbFormStore>((set) => ({
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
  initialValues?: DynamoDbFormValues;
  onSubmit?(values: DynamoDbFormValues): Promise<void>;
  canViewSecrets?: boolean;
  getValueWithSecrets?(): Promise<DynamoDbFormValues | undefined>;
  connectionId?: string;
}

export default function DynamoDbForm(props: Props): ReactElement<any> {
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
    return DynamoDbFormValues.validate(values, {
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
        <CheckConnectionButton
          isValid={Object.keys(errors).length === 0}
          getRequest={() => {
            return createMessage(CheckConnectionConfigRequestSchema, {
              connectionConfig: buildConnectionConfigDynamoDB({
                ...formData,
              }),
            });
          }}
          getRequestById={() => {
            return createMessage(CheckConnectionConfigByIdRequestSchema, {
              id: connectionId ?? '',
            });
          }}
          connectionName={formData.connectionName}
          connectionType="dynamodb"
          mode={mode === 'view' ? 'checkById' : 'check'}
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
