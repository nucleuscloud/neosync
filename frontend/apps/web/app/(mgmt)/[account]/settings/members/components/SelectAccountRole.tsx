import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getAccountRoleString } from '@/util/util';
import { AccountRole } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  role: AccountRole;
  onChange(role: AccountRole): void;
}

export default function SelectAccountRole(props: Props): ReactElement<any> {
  const { role, onChange } = props;

  return (
    <Select
      onValueChange={(newValue) => {
        if (newValue) {
          onChange(parseInt(newValue, 10) as AccountRole);
        }
      }}
      value={role.toString()}
    >
      <SelectTrigger>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {[
          AccountRole.ADMIN,
          AccountRole.JOB_DEVELOPER,
          AccountRole.JOB_EXECUTOR,
          AccountRole.JOB_VIEWER,
        ].map((role) => (
          <SelectItem
            key={role}
            className="cursor-pointer"
            value={role.toString()}
          >
            {getAccountRoleString(role)}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
