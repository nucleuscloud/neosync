import FormErrorMessage from '@/components/FormErrorMessage';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { NewJobHook } from '@neosync/sdk';
import { FormEvent, ReactElement } from 'react';
import { ValidationError } from 'yup';
import FormHeader from './FormHeader';
import JobConfigSqlForm from './JobConfigSqlForm';
import { useNewHookStore } from './stores';
import {
  HookTypeFormValue,
  newFormDataToNewJobHook,
  NewJobHookFormValues,
} from './validation';

interface Props {
  onSubmit(values: NewJobHook): Promise<void>;
  jobConnectionIds: string[];
}

export default function NewHookForm(props: Props): ReactElement {
  const { onSubmit, jobConnectionIds } = props;
  const {
    formData,
    errors,
    setFormData,
    setErrors,
    setSubmitting,
    isSubmitting,
  } = useNewHookStore();

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    if (isSubmitting) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await NewJobHookFormValues.validate(formData, {
        abortEarly: false,
      });
      await onSubmit(newFormDataToNewJobHook(formData));
    } catch (err) {
      if (err instanceof ValidationError) {
        const validationErrors: Record<string, string> = {};
        err.inner.forEach((error) => {
          if (error.path) {
            validationErrors[error.path] = error.message;
          }
        });
        setErrors(validationErrors);
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <FormHeader
          htmlFor="name"
          title="Name"
          description="Name of the hook for display and reference, must be unique"
          isErrored={!!errors.name}
        />
        <Input
          id="name"
          autoCapitalize="off" // we don't allow capitals in team names
          data-1p-ignore // tells 1password extension to not autofill this field
          value={formData.name || ''}
          onChange={(e) => setFormData({ name: e.target.value })}
          placeholder="Hook name"
        />
        <FormErrorMessage message={errors.name} />
      </div>

      <div className="space-y-2">
        <FormHeader
          htmlFor="description"
          title="Description"
          description="What this hook does"
          isErrored={!!errors.description}
        />
        <Textarea
          id="description"
          value={formData.description || ''}
          onChange={(e) => setFormData({ description: e.target.value })}
          placeholder="Hook description"
          rows={3}
        />
        <FormErrorMessage message={errors.description} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <FormHeader
            htmlFor="priority"
            title="Priority (0-100)"
            description="Determines execution order. Lower values are higher priority"
            isErrored={!!errors.priority}
          />
          <Input
            id="priority"
            type="number"
            value={formData.priority}
            onChange={(e) => setFormData({ priority: e.target.valueAsNumber })}
          />
          <FormErrorMessage message={errors.priority} />
        </div>
      </div>

      <div className="flex items-center gap-4">
        <FormHeader
          htmlFor="enabled"
          title="Enabled"
          description="Whether or not this hook will be invoked during a job run"
          isErrored={!!errors.enabled}
        />
        <Switch
          id="enabled"
          checked={formData.enabled || false}
          onCheckedChange={(checked) => setFormData({ enabled: checked })}
        />
        <FormErrorMessage message={errors.enabled} />
      </div>

      <div className="flex flex-col gap-4">
        <FormHeader
          title="Hook Type"
          description="The type of hook. Currently only SQL hooks are supported"
          isErrored={!!errors.hookType}
        />
        <ToggleGroup
          className="flex justify-start"
          type="single"
          onValueChange={(value) => {
            if (value) {
              setFormData({ hookType: value as HookTypeFormValue });
            }
          }}
          value={formData.hookType}
        >
          <ToggleGroupItem value="sql">SQL</ToggleGroupItem>
        </ToggleGroup>
        <FormErrorMessage message={errors.hookType} />
      </div>

      <div className="flex flex-col gap-4">
        {formData.hookType === 'sql' && (
          <JobConfigSqlForm
            values={formData.sql}
            setValues={(newSqlData) => {
              setFormData({ sql: newSqlData });
            }}
            jobConnectionIds={jobConnectionIds}
            errors={errors}
          />
        )}
      </div>

      <div className="flex justify-end gap-3">
        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full sm:w-auto"
        >
          {isSubmitting ? 'Submitting...' : 'Submit'}
        </Button>
      </div>
    </form>
  );
}
