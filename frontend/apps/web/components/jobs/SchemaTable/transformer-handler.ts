import {
  SystemTransformer,
  TransformerSource,
  UserDefinedTransformer,
} from '@neosync/sdk';

export interface BasicTransformerHandler {
  getSystemTransformers(): SystemTransformer[];
  getUserDefinedTransformers(): UserDefinedTransformer[];

  getSystemTransformerBySource(
    source: TransformerSource
  ): SystemTransformer | undefined;
  getUserDefinedTransformerById(id: string): UserDefinedTransformer | undefined;
}

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

  public getFilteredTransformers(filters: any): {
    system: SystemTransformer[];
    userDefined: UserDefinedTransformer[];
  } {
    return { system: [], userDefined: [] };
  }

  public getSystemTransformers(): SystemTransformer[] {
    return this.systemTransformers;
  }
  public getUserDefinedTransformers(): UserDefinedTransformer[] {
    return this.userDefinedTransformers;
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
