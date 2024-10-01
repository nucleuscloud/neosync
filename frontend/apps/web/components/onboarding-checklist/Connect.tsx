import { AwsS3Icon } from '@/public/icons/AwsS3Icon';
import { DynamoDBIcon } from '@/public/icons/DynamoDB';
import { NeonGrayScale } from '@/public/icons/NeonGrayScale';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement } from 'react';
import { DiMongodb } from 'react-icons/di';
import {
  SiGooglecloudstorage,
  SiMicrosoftsqlserver,
  SiPostgresql,
  SiSupabase,
} from 'react-icons/si';
import { MysqlIcon } from '../../public/icons/Mysql';
import { Button } from '../ui/button';

interface Props {
  onNextStep: () => void;
  onPreviousStep: () => void;
}

export default function Connect(props: Props): ReactElement {
  const { onPreviousStep, onNextStep } = props;

  const theme = useTheme();

  const integrations = [
    { name: 'Postgres', icon: <SiPostgresql className="w-8 h-8" /> },
    { name: 'Mysql', icon: <MysqlIcon theme={theme.resolvedTheme} /> },
    { name: 'Mongo DB', icon: <DiMongodb className="w-8 h-8" /> },
    { name: 'Dynamo DB', icon: <DynamoDBIcon theme={theme.resolvedTheme} /> },
    { name: 'SQL Server', icon: <SiMicrosoftsqlserver className="w-8 h-8" /> },
    { name: 'AWS S3', icon: <AwsS3Icon theme={theme.resolvedTheme} /> },
    { name: 'GCS', icon: <SiGooglecloudstorage className="w-8 h-8" /> },
    { name: 'Supabase', icon: <SiSupabase className="w-8  h-8" /> },
    { name: 'Neon', icon: <NeonGrayScale theme={theme.resolvedTheme} /> },
  ];

  return (
    <div className="flex flex-col gap-12 justify-center items-center text-center">
      <h1 className="font-semibold text-2xl">Connect</h1>
      <p className="text-sm px-10">
        Create <span className="font-semibold">Connections</span> to your
        upstream and downstream data sources. Neosync supports a variety of
        databases and types of object storage.
      </p>
      <div className="grid grid-cols-3 gap-2">
        {integrations.map((item) => (
          <div
            className="p-4 border border-gray-300 dark:border-[#0D47F0] flex flex-col items-center shadow-lg rounded-lg"
            key={item.name}
          >
            {item.icon}
            <span className="mt-2 text-xs">{item.name}</span>
          </div>
        ))}
      </div>

      <div className="flex flex-row justify-between w-full py-6">
        <Button variant="outline" type="reset" onClick={onPreviousStep}>
          Back
        </Button>
        <Button onClick={onNextStep}>
          <div className="flex flex-row items-center gap-2">
            <div>Next</div> <ArrowRightIcon />
          </div>
        </Button>
      </div>
    </div>
  );
}
