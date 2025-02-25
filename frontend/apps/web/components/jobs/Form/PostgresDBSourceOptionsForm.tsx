import { PostgresSourceOptionsFormValues } from '@/yup-validations/jobs';
import { ReactElement } from 'react';
import ColumnRemovalStrategyOptionsForm from './ColumnRemovalStrategyOptionsForm';
import NewColumnAdditionStrategyOptionsForm from './NewColumnAdditionStrategyOptionsForm';

interface Props {
  value: PostgresSourceOptionsFormValues;
  setValue(newVal: PostgresSourceOptionsFormValues): void;
}

export default function PostgresDBSourceOptionsForm(
  props: Props
): ReactElement<any> {
  const { value, setValue } = props;

  return (
    <div className="flex flex-col md:flex-row gap-6 pb-2">
      <div className="w-full">
        <NewColumnAdditionStrategyOptionsForm
          value={value.newColumnAdditionStrategy}
          setValue={(strategy) => {
            setValue({ ...value, newColumnAdditionStrategy: strategy });
          }}
        />
      </div>
      <div className="w-full">
        <ColumnRemovalStrategyOptionsForm
          value={value.columnRemovalStrategy}
          setValue={(strategy) => {
            setValue({ ...value, columnRemovalStrategy: strategy });
          }}
        />
      </div>
    </div>
  );
}
