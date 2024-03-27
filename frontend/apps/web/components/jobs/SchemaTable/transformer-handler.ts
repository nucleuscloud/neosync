import { SystemTransformer, UserDefinedTransformer } from '@neosync/sdk';

export class TransformerHandler {
  private readonly systemTransformers: SystemTransformer[];
  private readonly userDefinedTransformers: UserDefinedTransformer[];

  private readonly systemBySource: Map<string, SystemTransformer>;
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

  public getAllSystemTransformers(): SystemTransformer[] {
    return this.systemTransformers;
  }
  public getAllUserDefinedTransformers(): UserDefinedTransformer[] {
    return this.userDefinedTransformers;
  }

  public getSystemTransformerBySource(
    source: string
  ): SystemTransformer | undefined {
    return this.systemBySource.get(source);
  }

  public getUserDefinedById(id: string): UserDefinedTransformer | undefined {
    return this.userDefinedById.get(id);
  }
}
