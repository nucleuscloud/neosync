import { PostgresSourceOptionsFormValues } from '@/yup-validations/jobs';
import { ReactElement } from 'react';
import NewColumnAdditionStrategyOptionsForm from './NewColumnAdditionStrategyOptionsForm';

interface Props {
  value: PostgresSourceOptionsFormValues;
  setValue(newVal: PostgresSourceOptionsFormValues): void;
}

export default function PostgresDBSourceOptionsForm(
  props: Props
): ReactElement {
  const { value, setValue } = props;

  return (
    <div className="flex flex-col gap-2 py-2">
      <NewColumnAdditionStrategyOptionsForm
        value={value.newColumnAdditionStrategy}
        setValue={(strategy) => {
          setValue({ ...value, newColumnAdditionStrategy: strategy });
        }}
      />
    </div>
  );
}
