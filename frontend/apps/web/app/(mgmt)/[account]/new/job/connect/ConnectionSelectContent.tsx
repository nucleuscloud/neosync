import { SelectGroup, SelectItem, SelectLabel } from '@/components/ui/select';
import { Connection } from '@neosync/sdk';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';

interface Props {
  postgres?: Connection[];
  mysql?: Connection[];
  s3?: Connection[];
  openai?: Connection[];

  newConnectionValue: string;
}
export default function ConnectionSelectContent(props: Props): ReactElement {
  const {
    postgres = [],
    mysql = [],
    s3 = [],
    openai = [],
    newConnectionValue,
  } = props;
  return (
    <>
      {postgres.length > 0 && (
        <SelectGroup>
          <SelectLabel>Postgres</SelectLabel>
          {postgres.map((connection) => (
            <SelectItem
              className="cursor-pointer ml-2"
              key={connection.id}
              value={connection.id}
            >
              {connection.name}
            </SelectItem>
          ))}
        </SelectGroup>
      )}

      {mysql.length > 0 && (
        <SelectGroup>
          <SelectLabel>Mysql</SelectLabel>
          {mysql.map((connection) => (
            <SelectItem
              className="cursor-pointer ml-2"
              key={connection.id}
              value={connection.id}
            >
              {connection.name}
            </SelectItem>
          ))}
        </SelectGroup>
      )}
      {s3.length > 0 && (
        <SelectGroup>
          <SelectLabel>AWS S3</SelectLabel>
          {s3.map((connection) => (
            <SelectItem
              className="cursor-pointer ml-2"
              key={connection.id}
              value={connection.id}
            >
              {connection.name}
            </SelectItem>
          ))}
        </SelectGroup>
      )}
      {openai.length > 0 && (
        <SelectGroup>
          <SelectLabel>OpenAI</SelectLabel>
          {openai.map((connection) => (
            <SelectItem
              className="cursor-pointer ml-2"
              key={connection.id}
              value={connection.id}
            >
              {connection.name}
            </SelectItem>
          ))}
        </SelectGroup>
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
