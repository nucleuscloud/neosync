import pandas as pd
from sqlalchemy import create_engine

df = pd.read_csv("uci-adult2.csv")
engine = create_engine('postgresql://postgres:foofar@localhost:5434/nucleus?sslmode=disable')

df.to_sql('adult', engine, if_exists='replace', index=False)

