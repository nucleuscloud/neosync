import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { Input } from '@/components/ui/input';
import { GcpCloudStorageFormValues } from '@/yup-validations/connections';
import { ReactElement } from 'react';

interface Props {
  value: GcpCloudStorageFormValues['gcp'];
  onChange: (value: GcpCloudStorageFormValues['gcp']) => void;
  errors: Record<string, string>;
}

export default function Bucket(props: Props): ReactElement {
  const { value, onChange, errors } = props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="bucket"
          title="Bucket"
          description="The GCP Cloud Storage bucket"
          isErrored={!!errors['gcp.bucket']}
          isRequired={true}
        />
        <Input
          id="bucket"
          value={value.bucket || ''}
          onChange={(e) => onChange({ ...value, bucket: e.target.value })}
        />
        <FormErrorMessage message={errors['gcp.bucket']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="pathPrefix"
          title="Path Prefix"
          description="The path prefix for the bucket"
          isErrored={!!errors['gcp.pathPrefix']}
        />
        <Input
          id="pathPrefix"
          value={value.pathPrefix || ''}
          onChange={(e) => onChange({ ...value, pathPrefix: e.target.value })}
        />
        <FormErrorMessage message={errors['gcp.pathPrefix']} />
      </div>
    </>
  );
}
