from names_dataset import NameDataset, NameWrapper
import pickle

nd = NameDataset()

# print(nd.get_country_codes(alpha_2=True))

# english_speaking_and_european_countries = [
#     'AL', 'AT', 'BE', 'BG', 'CA', 'CH', 'CY', 'CZ', 'DE', 'DK', 'EE', 'ES', 'FI', 'FR', 'GB', 'GR', 'HR', 'HU', 'IE', 'IS', 'IT', 'LT', 'LU', 'LV', 'MT', 'NL', 'NO', 'PL', 'PT', 'RO', 'SE', 'SI', 'SK', 'US', 'IE', 'GB', 'JM', 'NG', 'ZA'
# ]

countries = [
  'US', 'GB', 'DE', 'FR', 'IT', 'NL', 'CA', 'AT', 'BE', 'BG', 'GR', 'IE', 'ES', 'SE'
]
all_first_names = []
all_last_names = []

for country in countries:
  print(f'processing country {country}...')
  top_first_names = nd.get_top_names(n=10000, use_first_names=True, country_alpha2=country)
  all_first_names += top_first_names[country]['M']
  all_first_names += top_first_names[country]['F']

  top_last_names = nd.get_top_names(n=10000, use_first_names=False, country_alpha2=country)
  all_last_names += top_last_names[country]
  all_last_names += top_last_names[country]

with open('first_names.pkl', 'wb') as f:
  pickle.dump(all_first_names, f)

with open('last_names.pkl', 'wb') as f:
  pickle.dump(all_last_names, f)
