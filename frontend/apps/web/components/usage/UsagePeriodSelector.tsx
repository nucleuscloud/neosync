import { ReactElement } from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import { UsagePeriod } from './util';

interface Props {
  period: UsagePeriod;
  setPeriod(newVal: UsagePeriod): void;
}

export default function UsagePeriodSelector(props: Props): ReactElement {
  const { period, setPeriod } = props;
  return (
    <Select
      onValueChange={(value: string) => {
        if (!value) {
          return;
        }
        const typedVal = value as UsagePeriod;
        setPeriod(typedVal);
      }}
      value={period}
    >
      <SelectTrigger>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem className="cursor-pointer" value="current">
          <p>Current Period</p>
        </SelectItem>
        <SelectItem className="cursor-pointer" value="last-month">
          <p>Last Month</p>
        </SelectItem>
      </SelectContent>
    </Select>
  );
}
