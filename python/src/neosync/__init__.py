import sys
import os

# Ensures import paths work from the root of the module
sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))

# Exports root files so they are importable from just "neosync" without having to specify the file name
from client import Neosync
