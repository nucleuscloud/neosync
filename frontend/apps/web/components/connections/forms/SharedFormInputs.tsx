import ButtonText from '@/components/ButtonText';
import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { PasswordInput } from '@/components/PasswordComponent';
import { PermissionConnectionType } from '@/components/permissions/columns';
import PermissionsDialog from '@/components/permissions/PermissionsDialog';
import { SecurePasswordInput } from '@/components/SecurePasswordInput';
import Spinner from '@/components/Spinner';
import SwitchCard from '@/components/switches/SwitchCard';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  AwsAdvancedFormValues,
  AwsCredentialsFormValues,
  ClientTlsFormValues,
  SqlOptionsFormValues,
  SshTunnelFormValues,
} from '@/yup-validations/connections';
import { create as createMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  CheckConnectionConfigRequest,
  CheckConnectionConfigResponse,
  CheckConnectionConfigResponseSchema,
  ConnectionService,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

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
          isErrored={!!errors['options.maxConnectionLimit']}
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
        <FormErrorMessage message={errors['options.maxConnectionLimit']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="maxOpenDuration"
          title="Max Open Duration"
          description="The maximum amount of time a connection may be reused. Expired connections may be closed laizly before reuse. Ex: 1s, 1m, 500ms. Empty to leave unset."
          isErrored={!!errors['options.maxOpenDuration']}
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
        <FormErrorMessage message={errors['options.maxOpenDuration']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="maxIdleLimit"
          title="Max Idle Limit"
          description="The maximum number of idle database connections allowed. If set to 0 then there is no limit on the number of idle connections. -1 to leave unset and use system default."
          isErrored={!!errors['options.maxIdleLimit']}
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
        <FormErrorMessage message={errors['options.maxIdleLimit']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="maxIdleDuration"
          title="Max Idle Duration"
          description="The maximum amount of time a connection may be idle. Expired connections may be closed laizly before reuse. Ex: 1s, 1m, 500ms. Empty to leave unset."
          isErrored={!!errors['options.maxIdleDuration']}
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
        <FormErrorMessage message={errors['options.maxIdleDuration']} />
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
          isErrored={!!errors['clientTls.rootCert']}
        />
        <Textarea
          id="rootCert"
          value={value.rootCert || ''}
          onChange={(e) => onChange({ ...value, rootCert: e.target.value })}
        />
        <FormErrorMessage message={errors['clientTls.rootCert']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="clientCert"
          title="Client Certificate"
          description={`A public key certificate issued to the client by a trusted
                      Certificate Authority (CA).`}
          isErrored={!!errors['clientTls.clientCert']}
        />
        <Textarea
          id="clientCert"
          value={value.clientCert || ''}
          onChange={(e) => onChange({ ...value, clientCert: e.target.value })}
        />
        <FormErrorMessage message={errors['clientTls.clientCert']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="clientKey"
          title="Client Key"
          description={`A private key corresponding to the client certificate.
                      The client key is used to authenticate the client to the
                      server.`}
          isErrored={!!errors['clientTls.clientKey']}
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
        <FormErrorMessage message={errors['clientTls.clientKey']} />
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
          isErrored={!!errors['clientTls.serverName']}
        />
        <Textarea
          id="serverName"
          value={value.serverName || ''}
          onChange={(e) => onChange({ ...value, serverName: e.target.value })}
        />
        <FormErrorMessage message={errors['clientTls.serverName']} />
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
          isErrored={!!errors['tunnel.host']}
        />
        <Input
          id="host"
          value={value.host || ''}
          onChange={(e) => onChange({ ...value, host: e.target.value })}
        />
        <FormErrorMessage message={errors['tunnel.host']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="port"
          title="Port"
          description="The port of the bastion host."
          isErrored={!!errors['tunnel.port']}
        />
        <Input
          id="port"
          type="number"
          value={value.port || ''}
          onChange={(e) => onChange({ ...value, port: e.target.valueAsNumber })}
        />
        <FormErrorMessage message={errors['tunnel.port']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="user"
          title="User"
          description="The name of the user that will be used to authenticate. If using passphrase auth, provide that in
                      the appropriate field below."
          isErrored={!!errors['tunnel.user']}
        />
        <Input
          id="user"
          value={value.user || ''}
          onChange={(e) => onChange({ ...value, user: e.target.value })}
        />
        <FormErrorMessage message={errors['tunnel.user']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="privateKey"
          title="Private Key"
          description={`The private key for the bastion host. If using passphrase auth, provide that in
                      the appropriate field below.`}
          isErrored={!!errors['tunnel.privateKey']}
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
        <FormErrorMessage message={errors['tunnel.privateKey']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="passphrase"
          title="Passphrase / Private Key Password"
          description="The passphrase that will be used to authenticate with. If
                      the SSH Key provided above is encrypted, provide the
                      password for it here."
          isErrored={!!errors['tunnel.passphrase']}
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
        <FormErrorMessage message={errors['tunnel.passphrase']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="knownHostPublicKey"
          title="Known Host Public Key"
          description={`The public key of the bastion host. This should be in the format
                      like what is found in the \`~/.ssh/known_hosts\` file,
                      excluding the hostname. If this is not provided, any host
                      public key will be accepted.`}
          isErrored={!!errors['tunnel.knownHostPublicKey']}
        />
        <Input
          id="knownHostPublicKey"
          value={value.knownHostPublicKey || ''}
          onChange={(e) =>
            onChange({ ...value, knownHostPublicKey: e.target.value })
          }
          placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAlkjd9s7aJkfdLk3jSLkfj2lk3j2lkfj2l3kjf2lkfj2l"
        />
        <FormErrorMessage message={errors['tunnel.knownHostPublicKey']} />
      </div>
    </>
  );
}

interface CheckConnectionButtonProps {
  isValid: boolean;

  getRequest(): CheckConnectionConfigRequest;
  connectionName: string;
  connectionType: PermissionConnectionType;
}

export function CheckConnectionButton(
  props: CheckConnectionButtonProps
): ReactElement {
  const { isValid, getRequest, connectionName, connectionType } = props;
  const [isChecking, setIsChecking] = useState(false);
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();
  const [openPermissionDialog, setOpenPermissionDialog] = useState(false);
  const { mutateAsync: checkConnectionConfig } = useMutation(
    ConnectionService.method.checkConnectionConfig
  );

  async function onClick(): Promise<void> {
    try {
      setIsChecking(true);
      const res = await checkConnectionConfig(getRequest());
      setValidationResponse(res);
      setOpenPermissionDialog(!!res?.isConnected);
    } catch (err) {
      setValidationResponse(
        createMessage(CheckConnectionConfigResponseSchema, {
          isConnected: false,
          connectionError: err instanceof Error ? err.message : 'unknown error',
        })
      );
    } finally {
      setIsChecking(false);
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <PermissionsDialog
        checkResponse={
          validationResponse ??
          createMessage(CheckConnectionConfigResponseSchema, {})
        }
        openPermissionDialog={openPermissionDialog}
        setOpenPermissionDialog={setOpenPermissionDialog}
        isValidating={isChecking}
        connectionName={connectionName}
        connectionType={connectionType}
      />
      <div className="flex justify-end">
        <Button
          variant="outline"
          disabled={!isValid}
          onClick={onClick}
          type="button"
        >
          <ButtonText
            leftIcon={isChecking ? <Spinner /> : undefined}
            text="Test Connection"
          />
        </Button>
      </div>
      {validationResponse && !validationResponse.isConnected && (
        <div>
          <ErrorAlert
            title="Unable to connect"
            description={
              validationResponse.connectionError ?? 'no error returned'
            }
          />
        </div>
      )}
    </div>
  );
}

interface ErrorAlertProps {
  title: string;
  description: string;
}

function ErrorAlert(props: ErrorAlertProps): ReactElement {
  const { title, description } = props;
  return (
    <Alert variant="destructive">
      <ExclamationTriangleIcon className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{description}</AlertDescription>
    </Alert>
  );
}

interface AwsAdvancedConfigAccordionProps {
  value: AwsAdvancedFormValues;
  onChange(value: AwsAdvancedFormValues): void;
  errors: Record<string, string>;
}

export function AwsAdvancedConfigAccordion(
  props: AwsAdvancedConfigAccordionProps
): ReactElement {
  return (
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value="advanced">
        <AccordionTrigger>AWS Advanced Configuration</AccordionTrigger>
        <AccordionContent className="flex flex-col gap-4 p-2">
          <p className="text-sm">
            This is an optional section and is used if you need to tweak the AWS
            SDK to connect to a different region or endpoint other than the
            default.
          </p>
          <AwsAdvancedConfig {...props} />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}

interface AwsAdvancedConfigProps {
  value: AwsAdvancedFormValues;
  onChange(value: AwsAdvancedFormValues): void;
  errors: Record<string, string>;
}

function AwsAdvancedConfig(props: AwsAdvancedConfigProps): ReactElement {
  const { value, onChange, errors } = props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="region"
          title="Region"
          description="The AWS region to target"
          isErrored={!!errors['advanced.region']}
        />
        <Input
          id="region"
          value={value.region || ''}
          onChange={(e) => onChange({ ...value, region: e.target.value })}
        />
        <FormErrorMessage message={errors['advanced.region']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="endpoint"
          title="Endpoint"
          description="The endpoint to target"
          isErrored={!!errors['advanced.endpoint']}
        />
        <Input
          id="endpoint"
          value={value.endpoint || ''}
          onChange={(e) => onChange({ ...value, endpoint: e.target.value })}
        />
        <FormErrorMessage message={errors['advanced.endpoint']} />
      </div>
    </>
  );
}

interface AwsCredentialsFormAccordionProps
  extends SecretRevealProps<AwsCredentialsFormValues> {
  value: AwsCredentialsFormValues;
  onChange(value: AwsCredentialsFormValues): void;
  errors: Record<string, string>;
}

export function AwsCredentialsFormAccordion(
  props: AwsCredentialsFormAccordionProps
): ReactElement {
  return (
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value="credentials">
        <AccordionTrigger>AWS Credentials</AccordionTrigger>
        <AccordionContent className="flex flex-col gap-4 p-2">
          <AwsCredentialsForm {...props} />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}

interface AwsCredentialsFormProps
  extends SecretRevealProps<AwsCredentialsFormValues> {
  value: AwsCredentialsFormValues;
  onChange(value: AwsCredentialsFormValues): void;
  errors: Record<string, string>;
}

function AwsCredentialsForm(props: AwsCredentialsFormProps): ReactElement {
  const { value, onChange, errors, isViewMode, canViewSecrets, onRevealClick } =
    props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="accessKeyId"
          title="Access Key ID"
          description="The AWS access key ID"
          isErrored={!!errors['credentials.accessKeyId']}
        />
        <Input
          id="accessKeyId"
          value={value.accessKeyId || ''}
          onChange={(e) => onChange({ ...value, accessKeyId: e.target.value })}
        />
        <FormErrorMessage message={errors['credentials.accessKeyId']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="secretAccessKey"
          title="Secret Access Key"
          description="The AWS secret access key"
          isErrored={!!errors['credentials.secretAccessKey']}
        />
        {isViewMode ? (
          <SecurePasswordInput
            value={value.secretAccessKey || ''}
            disabled={!canViewSecrets}
            onRevealPassword={
              canViewSecrets
                ? async () => {
                    const values = await onRevealClick();
                    return values?.secretAccessKey ?? '';
                  }
                : undefined
            }
          />
        ) : (
          <PasswordInput
            id="secretAccessKey"
            value={value.secretAccessKey || ''}
            onChange={(e) =>
              onChange({ ...value, secretAccessKey: e.target.value })
            }
          />
        )}
        <FormErrorMessage message={errors['credentials.secretAccessKey']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="sessionToken"
          title="Session Token"
          description="The AWS session token"
          isErrored={!!errors['credentials.sessionToken']}
        />
        <Input
          id="sessionToken"
          value={value.sessionToken || ''}
          onChange={(e) => onChange({ ...value, sessionToken: e.target.value })}
        />
        <FormErrorMessage message={errors['credentials.sessionToken']} />
      </div>
      <div className="space-y-2">
        <SwitchCard
          isChecked={value.fromEc2Role ?? false}
          title="From EC2 Role"
          description="If true, the SDK will use the EC2 role assigned to the instance"
          onCheckedChange={(checked) =>
            onChange({ ...value, fromEc2Role: checked })
          }
        />
        <FormErrorMessage message={errors['credentials.fromEc2Role']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="roleArn"
          title="Role ARN"
          description="The AWS role ARN"
          isErrored={!!errors['credentials.roleArn']}
        />
        <Input
          id="roleArn"
          value={value.roleArn || ''}
          onChange={(e) => onChange({ ...value, roleArn: e.target.value })}
        />
        <FormErrorMessage message={errors['credentials.roleArn']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="roleExternalId"
          title="Role External ID"
          description="The AWS role external ID"
          isErrored={!!errors['credentials.roleExternalId']}
        />
        <Input
          id="roleExternalId"
          value={value.roleExternalId || ''}
          onChange={(e) =>
            onChange({ ...value, roleExternalId: e.target.value })
          }
        />
        <FormErrorMessage message={errors['credentials.roleExternalId']} />
      </div>
    </>
  );
}
