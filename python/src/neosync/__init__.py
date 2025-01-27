import sys
import os

# Ensures import paths work from the root of the module
sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))

# Exports root files so they are importable from just "neosync" without having to specify the file name
from neosync.client import Neosync as Neosync

from neosync.mgmt.v1alpha1.anonymization_pb2 import *
from neosync.mgmt.v1alpha1.api_key_pb2 import *
from neosync.mgmt.v1alpha1.connection_data_pb2 import *
from neosync.mgmt.v1alpha1.connection_pb2 import *
from neosync.mgmt.v1alpha1.job_pb2 import *
from neosync.mgmt.v1alpha1.metrics_pb2 import *
from neosync.mgmt.v1alpha1.transformer_pb2 import *
from neosync.mgmt.v1alpha1.user_account_pb2 import *
