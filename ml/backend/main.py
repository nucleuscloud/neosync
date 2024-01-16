import os

from ctgan.synthesizer import CTGAN
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel


class Item(BaseModel):
   modelpath: str
   num_samples: int

app = FastAPI()

# @app.get("/")
# def root():
#   return {"Hello": "World"}

# @app.get("/items/{item_id}")
# def read_item(item_id: int, q: Union[str, None] = None):
#     return {"item_id": item_id, "q": q}

@app.post("/predict")
async def sample(item: Item):
  if item.modelpath == None:
    raise HTTPException(status_code=404, detail="Must provide model path")

  directory, filename = os.path.split(item.modelpath)
  model = CTGAN.load_from_dir_for_eval(directory, filename)
  samples = model.sample(item.num_samples)

  return {"samples" : samples}


  # model = torch.load(item.modelpath)
  # model.eval()

  # with torch.no_grad():
  #    prediction = model(item.input_data)
  # prediction = prediction.tolist()
  # return {"prediction": prediction}
