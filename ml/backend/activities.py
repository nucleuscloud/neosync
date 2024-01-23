import logging
import os
from dataclasses import dataclass
from typing import List

import pandas as pd
from ctgan.synthesizer import CTGAN
from sqlalchemy import create_engine
from temporalio import activity


@dataclass
class TrainModelInput:
  epochs: int
  discrete_columns: List[str]
  modelpath: str # where to save the model to

  dsn: str

  schema: str
  table: str
  columns: List[str]

@activity.defn(name="ctgan_single_table_train")
async def train_model(input: TrainModelInput):
  activity.logger.info("running train model activity")
  engine = create_engine(input.dsn)
  activity.logger.info("reading in sql")
  joined_cols = ", ".join(input.columns)
  df = pd.read_sql(f"SELECT {joined_cols} FROM {input.schema}.{input.table};", engine)

  activity.logger.info("sql loaded, fitting into CTGAN model")
  ctgan = CTGAN()
  ctgan.fit(df, epochs=input.epochs, discrete_columns=input.discrete_columns)
  activity.logger.info("CTGAN is now fit, saving to modelpath")

  dirname, filename = os.path.split(input.modelpath)
  ctgan.save(dirname, filename)
  activity.logger.info("train model activity is now complete")

@dataclass
class SampleModelInput:
  num_samples: int
  modelpath: str

  dsn: str
  schema: str
  table: str


@activity.defn
async def sample_model(input: SampleModelInput):
  activity.logger.info("running sample model activity")
  dirname, filename = os.path.split(input.modelpath)
  ctgan = CTGAN.load_from_dir_for_eval(dirname, filename)
  activity.logger.info("ctgan is loaded from modelpath and ready for evaluation")
  samples = ctgan.sample(input.num_samples)

  activity.logger.info("sampling is now complete, storing into destination engine")

  engine = create_engine(input.dsn)

  samples.to_sql(input.table, engine, schema=input.schema, if_exists='append', index=False)
  activity.logger.info("sample model activity is now complete")

