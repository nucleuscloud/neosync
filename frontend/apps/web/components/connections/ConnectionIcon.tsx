import {
  ConnectionConfigCase,
  ConnectionTypeVariant,
} from '@/app/(mgmt)/[account]/connections/util';
import { NeonLogo } from '@/app/(mgmt)/[account]/new/connection/neon/NeonLogo';
import { OpenAiLogo } from '@/app/(mgmt)/[account]/new/connection/openai/OpenAiLogo';
import { SupabaseLogo } from '@/app/(mgmt)/[account]/new/connection/supabase/SupabaseLogo';
import { MysqlIcon } from '@/public/icons/Mysql';
import { useTheme } from 'next-themes';
import { ReactElement } from 'react';
import { DiMongodb, DiMsqlServer, DiPostgresql } from 'react-icons/di';
import { FaAws } from 'react-icons/fa';
import { SiGooglecloud } from 'react-icons/si';

interface Props {
  connectionType: ConnectionConfigCase;
  connectionTypeVariant?: ConnectionTypeVariant;
  iconWidth?: string;
  iconHeight?: string;
}

export default function ConnectionIcon(props: Props): ReactElement | null {
  const { connectionType, connectionTypeVariant, iconWidth, iconHeight } =
    props;

  const width = iconWidth || '40px';
  const height = iconHeight || '40px';

  const { resolvedTheme } = useTheme();

  switch (connectionType) {
    case 'pgConfig': {
      switch (connectionTypeVariant) {
        case 'neon': {
          return <NeonLogo />;
        }
        case 'supabase': {
          return <SupabaseLogo />;
        }
      }
      return <DiPostgresql style={{ width, height }} />;
    }
    case 'mysqlConfig': {
      return <MysqlIcon theme={resolvedTheme} />;
    }
    case 'awsS3Config': {
      return <FaAws style={{ width, height }} />;
    }
    case 'dynamodbConfig': {
      return <FaAws style={{ width, height }} />;
    }
    case 'openaiConfig': {
      return <OpenAiLogo bg={resolvedTheme === 'dark' ? 'white' : '#272F30'} />;
    }
    case 'mongoConfig': {
      return <DiMongodb style={{ width, height }} />;
    }
    case 'gcpCloudstorageConfig': {
      return <SiGooglecloud style={{ width, height }} />;
    }
    case 'mssqlConfig': {
      return <DiMsqlServer style={{ width, height }} />;
    }

    default:
      return null;
  }
}
