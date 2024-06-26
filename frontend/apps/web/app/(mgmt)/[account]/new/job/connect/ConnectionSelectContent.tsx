import { SelectGroup, SelectItem, SelectLabel } from '@/components/ui/select';
import { Connection } from '@neosync/sdk';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';

interface Props {
  postgres?: Connection[];
  mysql?: Connection[];
  s3?: Connection[];
  openai?: Connection[];
  mongodb?: Connection[];
  gcpcs?: Connection[];

  newConnectionValue: string;
}
export default function ConnectionSelectContent(props: Props): ReactElement {
  const {
    postgres = [],
    mysql = [],
    s3 = [],
    openai = [],
    mongodb = [],
    gcpcs = [],
    newConnectionValue,
  } = props;
  const selectgroups = [
    [postgres, 'Postgres'],
    [mysql, 'Mysql'],
    [mongodb, 'MongoDB'],
    [s3, 'AWS S3'],
    [openai, 'OpenAI'],
    [gcpcs, 'GCP Cloud Storage'],
  ] as const;
  return (
    <>
      {selectgroups.map(
        ([connections, label]) =>
          connections.length > 0 && (
            <SelectGroup key={label}>
              <SelectLabel>{label}</SelectLabel>
              {connections.map((connection) => (
                <SelectItem
                  className="cursor-pointer ml-2"
                  key={connection.id}
                  value={connection.id}
                >
                  {connection.name}
                </SelectItem>
              ))}
            </SelectGroup>
          )
      )}
      <SelectItem
        className="cursor-pointer"
        key="new-dst-connection"
        value={newConnectionValue}
      >
        <div className="flex flex-row gap-1 items-center">
          <PlusIcon />
          <p>New Connection</p>
        </div>
      </SelectItem>
    </>
  );
}
