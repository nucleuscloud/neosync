import DualListBox, {
  EmptyStateMessage,
  Option,
} from '@/components/DualListBox/DualListBox';
import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { useAccount } from '@/components/providers/account-provider';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { splitConnections } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import { TableIcon } from '@radix-ui/react-icons';
import { ReactElement, useCallback, useMemo } from 'react';
import ConnectionSelectContent from '../../connect/ConnectionSelectContent';
import {
  DataSamplingFormValue,
  FilterPatternTableIdentifier,
  TableScanFilterModeFormValue,
  TableScanFilterPatternsFormValue,
} from '../../job-form-validations';

interface SourceConnectionIdProps {
  error?: string;
  value: string;
  onChange(value: string): void;
}

export function SourceConnectionId(
  props: SourceConnectionIdProps
): ReactElement {
  const { error, value, onChange } = props;

  const { account } = useAccount();
  const {
    data: connectionsResp,
    isLoading,
    isPending,
  } = useQuery(
    ConnectionService.method.getConnections,
    {
      accountId: account?.id,
    },
    { enabled: !!account?.id }
  );

  const connections = useMemo(() => {
    if (isLoading || isPending || !connectionsResp) {
      return { postgres: [], mysql: [], mssql: [] };
    }
    return splitConnections(connectionsResp.connections);
  }, [connectionsResp, isLoading, isPending]);

  return (
    <div className="flex flex-col gap-4">
      <FormHeader
        title="Connection"
        description="The connection to use for the PII detection job."
        isErrored={!!error}
        labelClassName="text-lg"
      />
      <Select
        value={value}
        onValueChange={(value) => {
          if (!value) {
            return;
          }
          onChange(value);
        }}
      >
        <SelectTrigger>
          <SelectValue placeholder="Select a source connection" />
        </SelectTrigger>
        <SelectContent>
          <ConnectionSelectContent
            postgres={connections.postgres}
            mysql={connections.mysql}
            mssql={connections.mssql}
          />
        </SelectContent>
      </Select>
      <FormErrorMessage message={error} />
    </div>
  );
}

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
        description="Allow the job to sample data from the source. If disabled, only the table DDLs will be used to detect PII. For more accurate results, enable data sampling."
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
        description="The mode to use for the table scan filter. This will determine what schemas and tables will be scanned for PII."
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
        description="The patterns to use for the table scan filter based on the mode selected. A combination of schemas and tables will be scanned based on the mode and selection."
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
        mode={mode}
      />
      <TableScanFilterPatternTables
        error={errors?.['patterns.tables']}
        value={value.tables}
        onChange={(newTables) => {
          onChange({ ...value, tables: newTables });
        }}
        availableTableIdentifiers={availableTableIdentifiers}
        mode={mode}
      />
    </div>
  );
}

interface TableScanFilterPatternSchemasProps {
  error?: string;
  value: string[];
  onChange(value: string[]): void;
  availableSchemas: string[];
  mode: TableScanFilterModeFormValue;
}

export function TableScanFilterPatternSchemas(
  props: TableScanFilterPatternSchemasProps
): ReactElement {
  const { error, value, onChange, availableSchemas, mode } = props;

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

  const leftEmptyStates = useGetSchemaLeftEmptyStates(mode);
  const rightEmptyStates = useGetSchemaRightEmptyStates(mode);
  const cardDescription = useSchemaCardDescription(mode);

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
          <CardDescription>{cardDescription}</CardDescription>
        </CardHeader>
        <CardContent>
          <DualListBox
            options={dualListBoxOpts}
            selected={selectedSchemas}
            onChange={onSelectedChange}
            mode={'many'}
            leftEmptyState={leftEmptyStates}
            rightEmptyState={rightEmptyStates}
          />
        </CardContent>
      </Card>
      <FormErrorMessage message={error} />
    </div>
  );
}

function useGetSchemaLeftEmptyStates(
  mode: TableScanFilterModeFormValue
): EmptyStateMessage {
  return useMemo(() => {
    switch (mode) {
      case 'include_all':
        return {
          noOptions: 'Unable to load schemas or found none',
          noSelected: 'All schemas have been added!',
        };
      case 'include':
        return {
          noOptions: 'Unable to load schemas or found none',
          noSelected: 'All schemas available have been included!',
        };
      case 'exclude':
        return {
          noOptions: 'Unable to load schemas or found none',
          noSelected: 'All schemas available have been excluded!',
        };
    }
  }, [mode]);
}

function useGetSchemaRightEmptyStates(
  mode: TableScanFilterModeFormValue
): EmptyStateMessage {
  return useMemo(() => {
    switch (mode) {
      case 'include_all':
        return {
          noOptions: 'Unable to load schemas or found none',
          noSelected: 'All schemas have been added!',
        };
      case 'include':
        return {
          noOptions: 'Unable to load schemas or found none',
          noSelected: 'Add schemas to scan for PII!',
        };
      case 'exclude':
        return {
          noOptions: 'Unable to load schemas or found none',
          noSelected: 'Add schemas to exclude from PII scanning!',
        };
    }
  }, [mode]);
}
function useSchemaCardDescription(mode: TableScanFilterModeFormValue): string {
  return useMemo(() => {
    switch (mode) {
      case 'include_all':
        return 'Select all schemas to scan for PII.';
      case 'include':
        return 'Select schemas to scan for PII.';
      case 'exclude':
        return 'Select schemas to exclude from PII scanning.';
    }
  }, [mode]);
}

interface TableScanFilterPatternTablesProps {
  error?: string;
  value: FilterPatternTableIdentifier[];
  onChange(value: FilterPatternTableIdentifier[]): void;

  availableTableIdentifiers: FilterPatternTableIdentifier[];
  mode: TableScanFilterModeFormValue;
}

export function TableScanFilterPatternTables(
  props: TableScanFilterPatternTablesProps
): ReactElement {
  const { error, value, onChange, availableTableIdentifiers, mode } = props;

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

  const leftEmptyStates = useGetTableLeftEmptyStates(mode);
  const rightEmptyStates = useGetTableRightEmptyStates(mode);
  const cardDescription = useTableCardDescription(mode);
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
          <CardDescription>{cardDescription}</CardDescription>
        </CardHeader>
        <CardContent>
          <DualListBox
            options={dualListBoxOpts}
            selected={selectedSchemas}
            onChange={onSelectedChange}
            mode={'many'}
            leftEmptyState={leftEmptyStates}
            rightEmptyState={rightEmptyStates}
          />
        </CardContent>
      </Card>
      <FormErrorMessage message={error} />
    </div>
  );
}

function useTableCardDescription(mode: TableScanFilterModeFormValue): string {
  return useMemo(() => {
    switch (mode) {
      case 'include_all':
        return 'Select all tables to scan for PII.';
      case 'include':
        return 'Select tables to scan for PII.';
      case 'exclude':
        return 'Select tables to exclude from PII scanning.';
    }
  }, [mode]);
}

function useGetTableLeftEmptyStates(
  mode: TableScanFilterModeFormValue
): EmptyStateMessage {
  return useMemo(() => {
    switch (mode) {
      case 'include_all':
        return {
          noOptions: 'Unable to load tables or found none',
          noSelected: 'All tables have been added!',
        };
      case 'include':
        return {
          noOptions: 'Unable to load tables or found none',
          noSelected: 'All tables available have been included!',
        };
      case 'exclude':
        return {
          noOptions: 'Unable to load tables or found none',
          noSelected: 'All tables available have been excluded!',
        };
    }
  }, [mode]);
}

function useGetTableRightEmptyStates(
  mode: TableScanFilterModeFormValue
): EmptyStateMessage {
  return useMemo(() => {
    switch (mode) {
      case 'include_all':
        return {
          noOptions: 'Unable to load tables or found none',
          noSelected: 'All tables have been added!',
        };
      case 'include':
        return {
          noOptions: 'Unable to load tables or found none',
          noSelected: 'Add tables to scan for PII!',
        };
      case 'exclude':
        return {
          noOptions: 'Unable to load tables or found none',
          noSelected: 'Add tables to exclude from PII scanning!',
        };
    }
  }, [mode]);
}
