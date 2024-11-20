import {
  SupportedJobType,
  SystemTransformer,
  TransformerConfig,
  TransformerDataType,
  TransformerSource,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { JobType } from './schema-constraint-handler';

// Helper function to extract the 'case' property from a config type
type ExtractCase<T> = T extends { case: infer U } ? U : never;

// Computed type that extracts all case types from the config union
export type TransformerConfigCase = NonNullable<
  ExtractCase<TransformerConfig['config']>
>;

export interface TransformerResult {
  system: SystemTransformer[];
  userDefined: UserDefinedTransformer[];
}

export class TransformerHandler {
  private readonly systemTransformers: SystemTransformer[];
  private readonly userDefinedTransformers: UserDefinedTransformer[];

  private readonly systemByType: Map<TransformerConfigCase, SystemTransformer>;
  // used by user-defined to go from User Defined -> System
  private readonly systemBySource: Map<TransformerSource, SystemTransformer>;
  private readonly userDefinedById: Map<string, UserDefinedTransformer>;

  constructor(
    systemTransformers: SystemTransformer[],
    userDefinedTransformers: UserDefinedTransformer[]
  ) {
    this.systemTransformers = systemTransformers;
    this.userDefinedTransformers = userDefinedTransformers;

    this.systemByType = new Map();
    systemTransformers.forEach((t) => {
      if (t.config?.config.case) {
        this.systemByType.set(t.config.config.case, t);
      }
    });

    this.systemBySource = new Map(systemTransformers.map((t) => [t.source, t]));
    this.userDefinedById = new Map(
      userDefinedTransformers.map((t) => [t.id, t])
    );
  }

  public getTransformers(): {
    system: SystemTransformer[];
    userDefined: UserDefinedTransformer[];
  } {
    return {
      system: this.systemTransformers,
      userDefined: this.userDefinedTransformers,
    };
  }

  public getFilteredTransformers(
    filters: TransformerFilters
  ): TransformerResult {
    const systemMap = new Map(
      this.systemTransformers.map((t) => [
        t.source,
        shouldIncludeSystem(t, filters),
      ])
    );

    const userMap = new Map(
      this.userDefinedTransformers.map((t) => {
        const underlyingSystem = this.systemBySource.get(t.source);
        if (!underlyingSystem) {
          // uh oh
          return [t.id, false];
        }
        return [t.id, systemMap.get(underlyingSystem.source) ?? false];
      })
    );
    return {
      system: this.systemTransformers.filter((t) => systemMap.get(t.source)),
      userDefined: this.userDefinedTransformers.filter((t) =>
        userMap.get(t.id)
      ),
    };
  }

  public getSystemTransformerByConfigCase(
    caseType: TransformerConfigCase | string
  ): SystemTransformer | undefined {
    return this.systemByType.get(caseType as TransformerConfigCase);
  }

  public getUserDefinedTransformerById(
    id: string
  ): UserDefinedTransformer | undefined {
    return this.userDefinedById.get(id);
  }
}

function shouldIncludeSystem(
  transformer: SystemTransformer,
  filters: TransformerFilters
): boolean {
  if (!transformer.supportedJobTypes.some((jt) => jt === filters.jobType)) {
    return false;
  }
  if (filters.isGenerated) {
    return transformer.source === TransformerSource.GENERATE_DEFAULT;
  }
  // if identitytype is 'a', which means always, no value may be provided other than the database default
  if (
    filters.identityType === 'a' ||
    filters.identityType === 'auto_increment' ||
    filters.identityType?.startsWith('IDENTITY')
  ) {
    return (
      transformer.source === TransformerSource.GENERATE_DEFAULT ||
      transformer.source == TransformerSource.PASSTHROUGH
    );
  }
  if (transformer.source === TransformerSource.GENERATE_DEFAULT) {
    return filters.hasDefault;
  }
  if (filters.isForeignKey || filters.isVirtualForeignKey) {
    const allowedFkTransformers = buildAllowedForeignKeyTransformers(filters);
    return allowedFkTransformers.some((t) => t === transformer.source);
  }
  if (filters.isNullable) {
    if (transformer.source === TransformerSource.GENERATE_NULL) {
      return true;
    }
    // if the current transformer does not support null, filter it out
    if (!transformer.dataTypes.some((dt) => dt === TransformerDataType.NULL)) {
      return false;
    }
  }
  const tdts = new Set(transformer.dataTypes);
  if (filters.dataType === TransformerDataType.UNSPECIFIED) {
    return tdts.has(TransformerDataType.ANY);
  }
  return tdts.has(filters.dataType) || tdts.has(TransformerDataType.ANY);
}

function buildAllowedForeignKeyTransformers(
  filters: TransformerFilters
): TransformerSource[] {
  const allowedFkTransformers = [];
  if (
    filters.jobType === SupportedJobType.UNSPECIFIED ||
    filters.jobType === SupportedJobType.SYNC
  ) {
    allowedFkTransformers.push(
      TransformerSource.PASSTHROUGH,
      TransformerSource.TRANSFORM_JAVASCRIPT
    );
  } else if (filters.jobType === SupportedJobType.GENERATE) {
    allowedFkTransformers.push(TransformerSource.GENERATE_JAVASCRIPT);
  }
  if (filters.isNullable) {
    allowedFkTransformers.push(TransformerSource.GENERATE_NULL);
  }
  if (filters.hasDefault) {
    allowedFkTransformers.push(TransformerSource.GENERATE_DEFAULT);
  }
  return allowedFkTransformers;
}

export interface TransformerFilters {
  isForeignKey: boolean;
  isVirtualForeignKey: boolean;
  dataType: TransformerDataType;
  isNullable: boolean;
  hasDefault: boolean;
  isGenerated: boolean;
  identityType?: string;
  jobType: SupportedJobType;
}

export function toSupportedJobtype(jobtype: JobType): SupportedJobType {
  if (jobtype === 'sync') {
    return SupportedJobType.SYNC;
  } else if (jobtype === 'generate') {
    return SupportedJobType.GENERATE;
  }
  return SupportedJobType.UNSPECIFIED;
}
