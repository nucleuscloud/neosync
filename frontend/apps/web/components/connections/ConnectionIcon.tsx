import {
  ConnectionConfigCase,
  ConnectionTypeVariant,
} from '@/app/(mgmt)/[account]/connections/util';
import { NeonLogo } from '@/app/(mgmt)/[account]/new/connection/neon/NeonLogo';
import { OpenAiLogo } from '@/app/(mgmt)/[account]/new/connection/openai/OpenAiLogo';
import { SupabaseLogo } from '@/app/(mgmt)/[account]/new/connection/supabase/SupabaseLogo';
import { useTheme } from 'next-themes';
import { ReactElement } from 'react';
import { IconContext } from 'react-icons';
import { DiMongodb, DiMsqlServer, DiMysql, DiPostgresql } from 'react-icons/di';
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
          return (
            <IconContext.Provider value={{ style: { width, height } }}>
              <NeonLogo />
            </IconContext.Provider>
          );
        }
        case 'supabase': {
          return (
            <IconContext.Provider value={{ style: { width, height } }}>
              <SupabaseLogo />
            </IconContext.Provider>
          );
        }
      }
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <DiPostgresql />
        </IconContext.Provider>
      );
    }
    case 'mysqlConfig': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <DiMysql />
        </IconContext.Provider>
      );
    }
    case 'awsS3Config':
    case 'dynamodbConfig': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <FaAws />
        </IconContext.Provider>
      );
    }
    case 'openaiConfig': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <OpenAiLogo bg={resolvedTheme === 'dark' ? 'white' : '#272F30'} />
        </IconContext.Provider>
      );
    }
    case 'mongoConfig': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <DiMongodb />
        </IconContext.Provider>
      );
    }
    case 'gcpCloudstorageConfig': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <SiGooglecloud />
        </IconContext.Provider>
      );
    }
    case 'mssqlConfig': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <DiMsqlServer />
        </IconContext.Provider>
      );
    }

    default:
      return null;
  }
}
