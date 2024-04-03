import {
  SupportedJobType,
  SystemTransformer,
  TransformerDataType,
  TransformerSource,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { JobType } from './schema-constraint-handler';

export class TransformerHandler {
  private readonly systemTransformers: SystemTransformer[];
  private readonly userDefinedTransformers: UserDefinedTransformer[];

  private readonly systemBySource: Map<TransformerSource, SystemTransformer>;
  private readonly userDefinedById: Map<string, UserDefinedTransformer>;

  constructor(
    systemTransformers: SystemTransformer[],
    userDefinedTransformers: UserDefinedTransformer[]
  ) {
    this.systemTransformers = systemTransformers;
    this.userDefinedTransformers = userDefinedTransformers;

    this.systemBySource = new Map(systemTransformers.map((t) => [t.source, t]));
    this.userDefinedById = new Map(
      userDefinedTransformers.map((t) => [t.id, t])
    );
  }

  public getFilteredTransformers(filters: TransformerFilters): {
    system: SystemTransformer[];
    userDefined: UserDefinedTransformer[];
  } {
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

  public getSystemTransformerBySource(
    source: TransformerSource
  ): SystemTransformer | undefined {
    return this.systemBySource.get(source);
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
  if (
    filters.hasDefault &&
    transformer.source === TransformerSource.GENERATE_DEFAULT
  ) {
    return true;
  }
  if (filters.isForeignKey) {
    if (filters.isNullable) {
      return (
        transformer.source === TransformerSource.GENERATE_NULL ||
        transformer.source === TransformerSource.PASSTHROUGH
      );
    }
    return transformer.source === TransformerSource.PASSTHROUGH;
  }
  if (!transformer.supportedJobTypes.some((jt) => jt === filters.jobType)) {
    return false;
  }
  if (
    filters.isNullable &&
    !transformer.dataTypes.some((dt) => dt === TransformerDataType.NULL)
  ) {
    return false;
  }
  const tdts = new Set(transformer.dataTypes);
  if (filters.dataType === TransformerDataType.UNSPECIFIED) {
    return tdts.has(TransformerDataType.ANY);
  }
  return tdts.has(filters.dataType) || tdts.has(TransformerDataType.ANY);
}

export interface TransformerFilters {
  isForeignKey: boolean;
  dataType: TransformerDataType;
  isNullable: boolean;
  hasDefault: boolean;
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
