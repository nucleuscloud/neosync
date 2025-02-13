import { OpenAiFormValues } from '@/yup-validations/connections';
import { ReactElement } from 'react';
import { Name } from '../SharedFormInputs';
import Sdk from './Sdk';

interface Props {
  values: OpenAiFormValues;
}

export default function ViewOpenAiForm(props: Props): ReactElement {
  const { values } = props;

  return (
    <div className="space-y-6">
      <Name
        error={undefined}
        value={values.connectionName}
        onChange={() => {}}
      />
      <Sdk errors={{}} value={values.sdk} onChange={() => {}} />
    </div>
  );
}
