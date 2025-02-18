import { buildConnectionConfigOpenAi } from '@/app/(mgmt)/[account]/connections/util';
import { OpenAiFormValues } from '@/yup-validations/connections';
import ConnectionForm, {
  CreateProps as BaseCreateProps,
  EditProps as BaseEditProps,
  ViewProps as BaseViewProps,
} from '../ConnectionForm';

import { ReactElement } from 'react';
import OpenAiForm from './OpenAiForm';

type CreateProps = Omit<
  BaseCreateProps<OpenAiFormValues>,
  'Form' | 'buildConnectionConfig'
>;
type EditProps = Omit<
  BaseEditProps<OpenAiFormValues>,
  'Form' | 'buildConnectionConfig'
>;
type ViewProps = Omit<
  BaseViewProps<OpenAiFormValues>,
  'Form' | 'buildConnectionConfig' | 'toFormValues'
>;

type Props = CreateProps | EditProps | ViewProps;

export default function OpenAiConnectionForm(props: Props): ReactElement {
  if (props.mode === 'view') {
    return (
      <ConnectionForm<OpenAiFormValues>
        {...props}
        Form={OpenAiForm}
        canViewSecrets={true}
        toFormValues={(connection) => {
          if (!connection.connectionConfig) {
            return undefined;
          }
          if (connection.connectionConfig.config.case !== 'openaiConfig') {
            return undefined;
          }
          return {
            connectionName: connection.name,
            sdk: {
              url: connection.connectionConfig.config.value.apiUrl,
              apiKey: connection.connectionConfig.config.value.apiKey,
            },
          };
        }}
      />
    );
  }
  return (
    <ConnectionForm<OpenAiFormValues>
      {...props}
      Form={OpenAiForm}
      buildConnectionConfig={buildConnectionConfigOpenAi}
    />
  );
}
