import pickle

with open('first_names.pkl', 'rb') as f:
  first_names = pickle.load(f)

with open('last_names.pkl', 'rb') as f:
  last_names = pickle.load(f)

first_names = list(set(first_names))
last_names = list(set(last_names))

first_names = [name for name in set(first_names) if name.isascii()]
last_names = [name for name in set(last_names) if name.isascii()]


first_names = sorted(first_names, key=lambda x: (len(x), x))
last_names = sorted(last_names, key=lambda x: (len(x), x))

with open("first_names.txt", "w") as file:
    # Loop through each item in the array
    for item in first_names:
        # Write each item to the file, followed by a newline character
        file.write(item + "\n")
with open("last_names.txt", "w") as file:
    # Loop through each item in the array
    for item in last_names:
        # Write each item to the file, followed by a newline character
        file.write(item + "\n")
