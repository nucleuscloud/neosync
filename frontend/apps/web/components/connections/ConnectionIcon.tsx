import { NeonLogo } from '@/app/(mgmt)/[account]/new/connection/neon/NeonLogo';
import { SupabaseLogo } from '@/app/(mgmt)/[account]/new/connection/supabase/SupabaseLogo';
import { ReactElement } from 'react';
import { IconContext } from 'react-icons';
import { DiMysql, DiPostgresql } from 'react-icons/di';
import { FaAws } from 'react-icons/fa';

interface Props {
  name: string;
  iconWidth?: string;
  iconHeight?: string;
}

export default function ConnectionIcon(props: Props): ReactElement | null {
  const { name, iconWidth, iconHeight } = props;

  const width = iconWidth || '40px';
  const height = iconHeight || '40px';

  switch (name.toLowerCase()) {
    case 'postgres': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <DiPostgresql />
        </IconContext.Provider>
      );
    }
    case 'mysql': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <DiMysql />
        </IconContext.Provider>
      );
    }
    case 'aws s3': {
      return (
        <IconContext.Provider value={{ style: { width, height } }}>
          <FaAws />
        </IconContext.Provider>
      );
    }
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

    default:
      return null;
  }
}
