import {
  ClientTlsFormValues,
  SqlOptionsFormValues,
  SshTunnelFormValues,
} from '@/yup-validations/connections';
import {
  ClientTlsConfig,
  SqlConnectionOptions,
  SSHAuthentication,
  SSHTunnel,
} from '@neosync/sdk';

export function getSqlOptionsFormValues(
  input: SqlConnectionOptions | undefined
): SqlOptionsFormValues {
  return {
    maxConnectionLimit: input?.maxConnectionLimit ?? 20,
    maxIdleDuration: input?.maxIdleDuration ?? '',
    maxIdleLimit: input?.maxIdleConnections ?? 2,
    maxOpenDuration: input?.maxOpenDuration ?? '',
  };
}

export function getClientTlsFormValues(
  input: ClientTlsConfig | undefined
): ClientTlsFormValues {
  return {
    rootCert: input?.rootCert ?? '',
    clientCert: input?.clientCert ?? '',
    clientKey: input?.clientKey ?? '',
    serverName: input?.serverName ?? '',
  };
}

export function getSshTunnelFormValues(
  input: SSHTunnel | undefined
): SshTunnelFormValues {
  return {
    host: input?.host ?? '',
    port: input?.port ?? 22,
    user: input?.user ?? '',
    privateKey: input?.authentication
      ? (getPrivateKeyFromSshAuthentication(input.authentication) ?? '')
      : '',
    passphrase: input?.authentication
      ? (getPassphraseFromSshAuthentication(input.authentication) ?? '')
      : '',
    knownHostPublicKey: input?.knownHostPublicKey ?? '',
  };
}

function getPassphraseFromSshAuthentication(
  sshauth: SSHAuthentication
): string | undefined {
  switch (sshauth.authConfig.case) {
    case 'passphrase':
      return sshauth.authConfig.value.value;
    case 'privateKey':
      return sshauth.authConfig.value.passphrase;
    default:
      return undefined;
  }
}

function getPrivateKeyFromSshAuthentication(
  sshauth: SSHAuthentication
): string | undefined {
  switch (sshauth.authConfig.case) {
    case 'privateKey':
      return sshauth.authConfig.value.value;
    default:
      return undefined;
  }
}
