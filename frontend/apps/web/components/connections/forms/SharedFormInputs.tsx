import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { SecurePasswordInput } from '@/components/SecurePasswordInput';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  ClientTlsFormValues,
  SqlOptionsFormValues,
  SshTunnelFormValues,
} from '@/yup-validations/connections';
import { ReactElement } from 'react';

export interface SecretRevealProps<T> {
  isViewMode: boolean;
  canViewSecrets: boolean;
  onRevealClick(): Promise<T | undefined>;
}

interface NameProps {
  error?: string;
  value: string;
  onChange(value: string): void;
}

export function Name(props: NameProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="name"
        title="Name"
        description="Name of the connection for display and reference, must be unique"
        isErrored={!!error}
      />
      <Input
        id="name"
        autoCapitalize="off" // we don't allow capitals
        data-1p-ignore // tells 1password extension to not autofill this field
        value={value || ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Connection name"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface SqlConnectionOptionsProps {
  value: SqlOptionsFormValues;
  onChange(value: SqlOptionsFormValues): void;
  errors: Record<string, string>;
}

export function SqlConnectionOptions(
  props: SqlConnectionOptionsProps
): ReactElement {
  const { value, onChange, errors } = props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="maxConnectionLimit"
          title="Max Connection Limit"
          description="The maximum number of concurrent database connections allowed. If set to 0 then there is no limit on the number of open connections. -1 to leave unset and use system default."
          isErrored={!!errors.maxConnectionLimit}
        />
        <Input
          id="maxConnectionLimit"
          className="max-w-[180px]"
          type="number"
          value={value.maxConnectionLimit || ''}
          onChange={(e) =>
            onChange({
              ...value,
              maxConnectionLimit: e.target.valueAsNumber,
            })
          }
        />
        <FormErrorMessage message={errors.maxConnectionLimit} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="maxOpenDuration"
          title="Max Open Duration"
          description="The maximum amount of time a connection may be reused. Expired connections may be closed laizly before reuse. Ex: 1s, 1m, 500ms. Empty to leave unset."
          isErrored={!!errors.maxOpenDuration}
        />
        <Input
          id="maxOpenDuration"
          className="max-w-[180px]"
          value={value.maxOpenDuration || ''}
          onChange={(e) =>
            onChange({
              ...value,
              maxOpenDuration: e.target.value,
            })
          }
        />
        <FormErrorMessage message={errors.maxOpenDuration} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="maxIdleLimit"
          title="Max Idle Limit"
          description="The maximum number of idle database connections allowed. If set to 0 then there is no limit on the number of idle connections. -1 to leave unset and use system default."
          isErrored={!!errors.maxIdleLimit}
        />
        <Input
          id="maxIdleLimit"
          className="max-w-[180px]"
          type="number"
          value={value.maxIdleLimit || ''}
          onChange={(e) =>
            onChange({
              ...value,
              maxIdleLimit: e.target.valueAsNumber,
            })
          }
        />
        <FormErrorMessage message={errors.maxIdleLimit} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="maxIdleDuration"
          title="Max Idle Duration"
          description="The maximum amount of time a connection may be idle. Expired connections may be closed laizly before reuse. Ex: 1s, 1m, 500ms. Empty to leave unset."
          isErrored={!!errors.maxIdleDuration}
        />
        <Input
          id="maxIdleDuration"
          className="max-w-[180px]"
          value={value.maxIdleDuration || ''}
          onChange={(e) =>
            onChange({
              ...value,
              maxIdleDuration: e.target.value,
            })
          }
        />
        <FormErrorMessage message={errors.maxIdleDuration} />
      </div>
    </>
  );
}

interface ClientTlsAccordionProps
  extends SecretRevealProps<ClientTlsFormValues> {
  value: ClientTlsFormValues;
  onChange(value: ClientTlsFormValues): void;
  errors: Record<string, string>;
}

export function ClientTlsAccordion(
  props: ClientTlsAccordionProps
): ReactElement {
  return (
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value="client-tls">
        <AccordionTrigger>Client TLS Certificates</AccordionTrigger>
        <AccordionContent className="flex flex-col gap-4 p-2">
          <div className="text-sm">
            Configuring this section allows Neosync to connect to the database
            using SSL/TLS.
          </div>
          <ClientTls {...props} />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}

interface ClientTlsProps extends SecretRevealProps<ClientTlsFormValues> {
  value: ClientTlsFormValues;
  onChange(value: ClientTlsFormValues): void;
  errors: Record<string, string>;
}

export function ClientTls(props: ClientTlsProps): ReactElement {
  const { value, onChange, errors, isViewMode, canViewSecrets, onRevealClick } =
    props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="rootCert"
          title="Root Certificate"
          description={`The public key certificate of the CA that issued the
                      server's certificate. Root certificates are used to
                      authenticate the server to the client. They ensure that
                      the server the client is connecting to is trusted.`}
          isErrored={!!errors.rootCert}
        />
        <Textarea
          id="rootCert"
          value={value.rootCert || ''}
          onChange={(e) => onChange({ ...value, rootCert: e.target.value })}
        />
        <FormErrorMessage message={errors.rootCert} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="clientCert"
          title="Client Certificate"
          description={`A public key certificate issued to the client by a trusted
                      Certificate Authority (CA).`}
          isErrored={!!errors.clientCert}
        />
        <Textarea
          id="clientCert"
          value={value.clientCert || ''}
          onChange={(e) => onChange({ ...value, clientCert: e.target.value })}
        />
        <FormErrorMessage message={errors.clientCert} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="clientKey"
          title="Client Key"
          description={`A private key corresponding to the client certificate.
                      The client key is used to authenticate the client to the
                      server.`}
          isErrored={!!errors.clientKey}
        />
        {isViewMode ? (
          <SecurePasswordInput
            value={value.clientKey || ''}
            disabled={!canViewSecrets}
            onRevealPassword={
              canViewSecrets
                ? async () => {
                    const values = await onRevealClick();
                    return values?.clientKey ?? '';
                  }
                : undefined
            }
          />
        ) : (
          <Textarea
            id="clientKey"
            value={value.clientKey || ''}
            onChange={(e) => onChange({ ...value, clientKey: e.target.value })}
          />
        )}
        <FormErrorMessage message={errors.clientKey} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="serverName"
          title="Server Name"
          description={`Server Name is used to verify the hostname on the returned
                      certificates. It is also included in the client's
                      handshake to support virtual hosting unless it is an IP
                      address. This is only required if performing full tls
                      verification.`}
          isErrored={!!errors.serverName}
        />
        <Textarea
          id="serverName"
          value={value.serverName || ''}
          onChange={(e) => onChange({ ...value, serverName: e.target.value })}
        />
        <FormErrorMessage message={errors.serverName} />
      </div>
    </>
  );
}

interface SshTunnelAccordionProps
  extends SecretRevealProps<SshTunnelFormValues> {
  value: SshTunnelFormValues;
  onChange(value: SshTunnelFormValues): void;
  errors: Record<string, string>;
}

export function SshTunnelAccordion(
  props: SshTunnelAccordionProps
): ReactElement {
  return (
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value="ssh-tunnel">
        <AccordionTrigger>SSH Tunnel</AccordionTrigger>
        <AccordionContent className="flex flex-col gap-4 p-2">
          <div className="text-sm">
            This section is optional and only necessary if your database is not
            publicly accessible to the internet.
          </div>
          <SSHTunnel {...props} />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}

interface SSHTunnelProps extends SecretRevealProps<SshTunnelFormValues> {
  value: SshTunnelFormValues;
  onChange(value: SshTunnelFormValues): void;
  errors: Record<string, string>;
}

export function SSHTunnel(props: SSHTunnelProps): ReactElement {
  const { value, onChange, errors, isViewMode, canViewSecrets, onRevealClick } =
    props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="host"
          title="Host"
          description="The hostname of the bastion host that will be used for SSH tunneling."
          isErrored={!!errors.host}
        />
        <Input
          id="host"
          value={value.host || ''}
          onChange={(e) => onChange({ ...value, host: e.target.value })}
        />
        <FormErrorMessage message={errors.host} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="port"
          title="Port"
          description="The port of the bastion host."
          isErrored={!!errors.port}
        />
        <Input
          id="port"
          value={value.port || ''}
          onChange={(e) => onChange({ ...value, port: e.target.valueAsNumber })}
        />
        <FormErrorMessage message={errors.port} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="user"
          title="User"
          description="The name of the user that will be used to authenticate. If using passphrase auth, provide that in
                      the appropriate field below."
          isErrored={!!errors.user}
        />
        <Input
          id="user"
          value={value.user || ''}
          onChange={(e) => onChange({ ...value, user: e.target.value })}
        />
        <FormErrorMessage message={errors.user} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="privateKey"
          title="Private Key"
          description={`The private key for the bastion host. If using passphrase auth, provide that in
                      the appropriate field below.`}
          isErrored={!!errors.privateKey}
        />
        {isViewMode ? (
          <SecurePasswordInput
            value={value.privateKey || ''}
            disabled={!canViewSecrets}
            onRevealPassword={
              canViewSecrets
                ? async () => {
                    const values = await onRevealClick();
                    return values?.privateKey ?? '';
                  }
                : undefined
            }
          />
        ) : (
          <Textarea
            id="privateKey"
            value={value.privateKey || ''}
            onChange={(e) => onChange({ ...value, privateKey: e.target.value })}
          />
        )}
        <FormErrorMessage message={errors.privateKey} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="passphrase"
          title="Passphrase / Private Key Password"
          description="The passphrase that will be used to authenticate with. If
                      the SSH Key provided above is encrypted, provide the
                      password for it here."
          isErrored={!!errors.passphrase}
        />
        {isViewMode ? (
          <SecurePasswordInput
            value={value.passphrase || ''}
            disabled={!canViewSecrets}
            onRevealPassword={
              canViewSecrets
                ? async () => {
                    const values = await onRevealClick();
                    return values?.passphrase ?? '';
                  }
                : undefined
            }
          />
        ) : (
          <Input
            id="passphrase"
            value={value.passphrase || ''}
            onChange={(e) => onChange({ ...value, passphrase: e.target.value })}
          />
        )}
        <FormErrorMessage message={errors.passphrase} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="knownHostPublicKey"
          title="Known Host Public Key"
          description={`The public key of the bastion host. This should be in the format
                      like what is found in the \`~/.ssh/known_hosts\` file,
                      excluding the hostname. If this is not provided, any host
                      public key will be accepted.`}
          isErrored={!!errors.knownHostPublicKey}
        />
        <Input
          id="knownHostPublicKey"
          value={value.knownHostPublicKey || ''}
          onChange={(e) =>
            onChange({ ...value, knownHostPublicKey: e.target.value })
          }
          placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAlkjd9s7aJkfdLk3jSLkfj2lk3j2lkfj2l3kjf2lkfj2l"
        />
        <FormErrorMessage message={errors.knownHostPublicKey} />
      </div>
    </>
  );
}
