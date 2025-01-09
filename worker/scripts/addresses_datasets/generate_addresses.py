import json
import os

with open('addresses.json', 'r') as f:
  addresses = json.load(f)

  # Extract unique values for each property and sort
  address1_values = sorted(set(addr['Address1'] for addr in addresses))
  address2_values = sorted(set(addr['Address2'] for addr in addresses))
  cities = sorted(set(addr['City'] for addr in addresses))
  states = sorted(set(addr['State'] for addr in addresses))
  zipcodes = sorted(set(addr['Zipcode'] for addr in addresses))

  # Create output directory if it doesn't exist
  os.makedirs('out', exist_ok=True)

  # Write each property to its own file
  with open('out/Address1.txt', 'w') as f:
    f.write('\n'.join(address1_values))

  with open('out/Address2.txt', 'w') as f:
    f.write('\n'.join(address2_values))

  with open('out/City.txt', 'w') as f:
    f.write('\n'.join(cities))

  with open('out/State.txt', 'w') as f:
    f.write('\n'.join(states))

  with open('out/Zipcode.txt', 'w') as f:
    f.write('\n'.join(zipcodes))
