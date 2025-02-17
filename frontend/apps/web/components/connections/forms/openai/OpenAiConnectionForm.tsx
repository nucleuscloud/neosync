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
  'ConnectionForm' | 'buildConnectionConfig'
>;
type EditProps = Omit<
  BaseEditProps<OpenAiFormValues>,
  'ConnectionForm' | 'buildConnectionConfig'
>;
type ViewProps = Omit<
  BaseViewProps<OpenAiFormValues>,
  'ConnectionForm' | 'buildConnectionConfig'
>;

type Props = CreateProps | EditProps | ViewProps;

export default function OpenAiConnectionForm(props: Props): ReactElement {
  if (props.mode === 'view') {
    return (
      <ConnectionForm<OpenAiFormValues>
        {...props}
        ConnectionForm={OpenAiForm}
      />
    );
  }
  return (
    <ConnectionForm<OpenAiFormValues>
      {...props}
      ConnectionForm={OpenAiForm}
      buildConnectionConfig={buildConnectionConfigOpenAi}
    />
  );
}
