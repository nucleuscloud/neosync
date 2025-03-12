import { ToggleGroupItem } from '@/components/ui/toggle-group';

import DualListBox, { Option } from '@/components/DualListBox/DualListBox';
import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import { ToggleGroup } from '@/components/ui/toggle-group';
import { TableIcon } from '@radix-ui/react-icons';
import { ReactElement, useCallback, useMemo } from 'react';
import {
  DataSamplingFormValue,
  FilterPatternTableIdentifier,
  TableScanFilterModeFormValue,
  TableScanFilterPatternsFormValue,
} from '../../job-form-validations';

interface UserPromptProps {
  error?: string;
  value: string;
  onChange(value: string): void;
}

export function UserPrompt(props: UserPromptProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="flex flex-col gap-4">
      <FormHeader
        title="User Prompt"
        description="Optionally input a prompt to guide the LLM part of the PII detection job."
        isErrored={!!error}
        labelClassName="text-lg"
      />
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Example: Non-PII columns: created_at, updated_at"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface DataSamplingProps {
  errors?: Record<string, string>;
  value: DataSamplingFormValue;
  onChange(value: DataSamplingFormValue): void;
}

export function DataSampling(props: DataSamplingProps): ReactElement {
  const { errors, value, onChange } = props;

  return (
    <div className="flex flex-col gap-4">
      <FormHeader
        title="Data Sampling"
        description="Allow the job to sample data from the source."
        isErrored={!!errors?.['isEnabled']}
        labelClassName="text-lg"
      />
      <ToggleGroup
        className="flex justify-start"
        type="single"
        onValueChange={(value) => {
          onChange({ isEnabled: value === 'enabled' });
        }}
        value={value.isEnabled ? 'enabled' : 'disabled'}
      >
        <ToggleGroupItem value="enabled">Enabled</ToggleGroupItem>
        <ToggleGroupItem value="disabled">Disabled</ToggleGroupItem>
      </ToggleGroup>
      <FormErrorMessage message={errors?.['isEnabled']} />
    </div>
  );
}

interface TableScanFilterModeProps {
  error?: string;
  value: TableScanFilterModeFormValue;
  onChange(value: TableScanFilterModeFormValue): void;
}

export function TableScanFilterMode(
  props: TableScanFilterModeProps
): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="flex flex-col gap-4">
      <FormHeader
        title="Table Scan Mode"
        description="The mode to use for the table scan filter"
        isErrored={!!error}
        labelClassName="text-lg"
      />
      <ToggleGroup
        className="flex justify-start"
        type="single"
        onValueChange={(value) => {
          if (value) {
            onChange(value as TableScanFilterModeFormValue);
          }
        }}
        value={value}
      >
        <ToggleGroupItem value="include_all">Include All</ToggleGroupItem>
        <ToggleGroupItem value="include">Include</ToggleGroupItem>
        <ToggleGroupItem value="exclude">Exclude</ToggleGroupItem>
      </ToggleGroup>
      <FormErrorMessage message={error} />
    </div>
  );
}

interface TableScanFilterPatternsProps {
  errors?: Record<string, string>;
  value: TableScanFilterPatternsFormValue;
  onChange(value: TableScanFilterPatternsFormValue): void;
  availableSchemas: string[];
  availableTableIdentifiers: FilterPatternTableIdentifier[];
  mode: TableScanFilterModeFormValue;
}

export function TableScanFilterPatterns(
  props: TableScanFilterPatternsProps
): ReactElement | null {
  const {
    errors,
    value,
    onChange,
    availableSchemas,
    availableTableIdentifiers,
    mode,
  } = props;

  if (mode === 'include_all') {
    return null;
  }

  return (
    <div className="flex flex-col gap-4">
      <FormHeader
        title="Patterns"
        description="The patterns to use for the table scan filter"
        isErrored={
          !!errors?.['patterns.schemas'] || !!errors?.['patterns.tables']
        }
        labelClassName="text-lg"
      />
      <TableScanFilterPatternSchemas
        error={errors?.['patterns.schemas']}
        value={value.schemas}
        onChange={(newSchemas) => {
          onChange({ ...value, schemas: newSchemas });
        }}
        availableSchemas={availableSchemas}
      />
      <TableScanFilterPatternTables
        error={errors?.['patterns.tables']}
        value={value.tables}
        onChange={(newTables) => {
          onChange({ ...value, tables: newTables });
        }}
        availableTableIdentifiers={availableTableIdentifiers}
      />
    </div>
  );
}

interface TableScanFilterPatternSchemasProps {
  error?: string;
  value: string[];
  onChange(value: string[]): void;
  availableSchemas: string[];
}

export function TableScanFilterPatternSchemas(
  props: TableScanFilterPatternSchemasProps
): ReactElement {
  const { error, value, onChange, availableSchemas } = props;

  const dualListBoxOpts = useMemo((): Option[] => {
    return availableSchemas.map((schema) => ({
      value: schema,
    }));
  }, [availableSchemas]);

  const selectedSchemas = useMemo((): Set<string> => {
    return new Set(value);
  }, [value]);

  const onSelectedChange = useCallback(
    (value: Set<string>) => {
      onChange(Array.from(value));
    },
    [onChange]
  );

  return (
    <div className="flex flex-col md:flex-row gap-3">
      <Card className="w-full">
        <CardHeader className="flex flex-col gap-2">
          <div className="flex flex-row items-center gap-2">
            <div className="flex">
              <TableIcon className="h-4 w-4" />
            </div>
            <CardTitle>Schema Selection</CardTitle>
          </div>
          <CardDescription>Select schemas to scan for PII.</CardDescription>
        </CardHeader>
        <CardContent>
          <DualListBox
            options={dualListBoxOpts}
            selected={selectedSchemas}
            onChange={onSelectedChange}
            mode={'many'}
            leftEmptyState={{
              noOptions: 'Unable to load schemas or found none',
              noSelected: 'All schemas have been added!',
            }}
            rightEmptyState={{
              noOptions: 'Unable to load schemas or found none',
              noSelected: 'Add schemas to scan for PII!',
            }}
          />
        </CardContent>
      </Card>
      <FormErrorMessage message={error} />
    </div>
  );
}

interface TableScanFilterPatternTablesProps {
  error?: string;
  value: FilterPatternTableIdentifier[];
  onChange(value: FilterPatternTableIdentifier[]): void;

  availableTableIdentifiers: FilterPatternTableIdentifier[];
}

export function TableScanFilterPatternTables(
  props: TableScanFilterPatternTablesProps
): ReactElement {
  const { error, value, onChange, availableTableIdentifiers } = props;

  const dualListBoxOpts = useMemo((): Option[] => {
    return availableTableIdentifiers.map((tableIdentifier) => ({
      value: `${tableIdentifier.schema}.${tableIdentifier.table}`,
    }));
  }, [availableTableIdentifiers]);

  const selectedSchemas = useMemo((): Set<string> => {
    return new Set(
      value.map(
        (tableIdentifier) =>
          `${tableIdentifier.schema}.${tableIdentifier.table}`
      )
    );
  }, [value]);

  const onSelectedChange = useCallback(
    (value: Set<string>) => {
      onChange(
        Array.from(value).map((tableIdentifier) => {
          const [schema, table] = tableIdentifier.split('.');
          return { schema, table };
        })
      );
    },
    [onChange]
  );

  return (
    <div className="flex flex-col md:flex-row gap-3">
      <Card className="w-full">
        <CardHeader className="flex flex-col gap-2">
          <div className="flex flex-row items-center gap-2">
            <div className="flex">
              <TableIcon className="h-4 w-4" />
            </div>
            <CardTitle>Table Selection</CardTitle>
          </div>
          <CardDescription>Select tables to scan for PII.</CardDescription>
        </CardHeader>
        <CardContent>
          <DualListBox
            options={dualListBoxOpts}
            selected={selectedSchemas}
            onChange={onSelectedChange}
            mode={'many'}
            leftEmptyState={{
              noOptions: 'Unable to load tables or found none',
              noSelected: 'All tables have been added!',
            }}
            rightEmptyState={{
              noOptions: 'Unable to load tables or found none',
              noSelected: 'Add tables to scan for PII!',
            }}
          />
        </CardContent>
      </Card>
      <FormErrorMessage message={error} />
    </div>
  );
}
