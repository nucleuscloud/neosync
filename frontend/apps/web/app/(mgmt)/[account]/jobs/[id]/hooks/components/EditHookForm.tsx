import { FormEvent, ReactElement, useEffect } from 'react';
import * as yup from 'yup';
import { create } from 'zustand';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { JobHook } from '@neosync/sdk';

interface EditHookStore {
  formData: Partial<JobHook>;
  errors: Record<string, string>;
  isSubmitting: boolean;
  setFormData: (data: Partial<JobHook>) => void;
  setErrors: (errors: Record<string, string>) => void;
  setSubmitting: (isSubmitting: boolean) => void;
  resetForm: () => void;
}

export const useEditHookStore = create<EditHookStore>((set) => ({
  formData: {},
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () => set({ formData: {}, errors: {}, isSubmitting: false }),
}));

const jobHookSqlSchema = yup.object().shape({
  query: yup.string().required('SQL query is required'),
  connectionId: yup.string().required('Connection ID is required'),
  timing: yup
    .object()
    .shape({
      timing: yup.object().shape({
        case: yup
          .string()
          .oneOf(['preSync', 'postSync'], 'Must select pre-sync or post-sync')
          .required('Timing is required'),
        value: yup.object(),
      }),
    })
    .required('Timing configuration is required'),
});

const jobHookConfigSchema = yup.object().shape({
  config: yup
    .object()
    .shape({
      case: yup
        .string()
        .oneOf(['sql'], 'Only SQL hooks are supported currently')
        .required('Hook type is required'),
      value: yup.lazy((value) => {
        if (value && value.case === 'sql') {
          return jobHookSqlSchema;
        }
        return yup.object();
      }),
    })
    .required('Configuration is required'),
});

export const jobHookSchema = yup.object().shape({
  name: yup
    .string()
    .required('Name is required')
    .min(3, 'Name must be at least 3 characters'),
  description: yup.string(),
  priority: yup
    .number()
    .required('Priority is required')
    .min(0, 'Priority must be between 0 and 100')
    .max(100, 'Priority must be between 0 and 100'),
  enabled: yup.boolean(),
  config: jobHookConfigSchema,
});

interface EditHookFormProps {
  hook: JobHook;
  onSubmit: (values: Partial<JobHook>) => Promise<void>;
}

export function EditHookForm({ hook, onSubmit }: EditHookFormProps) {
  const {
    formData,
    errors,
    setFormData,
    setErrors,
    setSubmitting,
    resetForm,
    isSubmitting,
  } = useEditHookStore();

  useEffect(() => {
    // Initialize form with hook data
    setFormData(hook);
    return () => resetForm();
  }, [hook, setFormData, resetForm]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    try {
      setSubmitting(true);
      setErrors({});

      // Validate form data
      const validatedData = await jobHookSchema.validate(formData, {
        abortEarly: false,
      });

      await onSubmit(validatedData);
    } catch (err) {
      if (err instanceof yup.ValidationError) {
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
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          autoCapitalize="off" // we don't allow capitals in team names
          data-1p-ignore // tells 1password extension to not autofill this field
          value={formData.name || ''}
          onChange={(e) => setFormData({ name: e.target.value })}
          placeholder="Hook name"
        />
        {errors.name && <p className="text-sm text-red-500">{errors.name}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          value={formData.description || ''}
          onChange={(e) => setFormData({ description: e.target.value })}
          placeholder="Hook description"
          rows={3}
        />
        {errors.description && (
          <p className="text-sm text-red-500">{errors.description}</p>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="priority">Priority (0-100)</Label>
          <Input
            id="priority"
            type="number"
            value={formData.priority}
            onChange={(e) => setFormData({ priority: e.target.valueAsNumber })}
          />
          {errors.priority && (
            <p className="text-sm text-red-500">{errors.priority}</p>
          )}
        </div>
      </div>

      <div className="flex items-center gap-4">
        <Label htmlFor="enabled">Enabled</Label>
        <Switch
          id="enabled"
          checked={formData.enabled || false}
          onCheckedChange={(checked) => setFormData({ enabled: checked })}
        />
      </div>

      <div className="flex justify-end gap-3">
        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full sm:w-auto"
        >
          {isSubmitting ? 'Saving...' : 'Save Changes'}
        </Button>
      </div>
    </form>
  );
}

interface JobConfigFormProps {}
function JobConfigForm(props: JobConfigFormProps): ReactElement {
  const {} = props;
}
