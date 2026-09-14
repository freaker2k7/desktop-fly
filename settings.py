import os

from dotenv import load_dotenv

load_dotenv()

# See: https://neuprint.janelia.org/account
TOKEN = os.environ.get("NEUPRINT_TOKEN")

if not TOKEN:
    raise ValueError("NEUPRINT_TOKEN environment variable is not set")

SERVER = "https://neuprint.janelia.org"
DATASET = "male-cns:v1.0"

ROOT_NEURON = "DNge104"

MAX_NEURONS = 250
MAX_CONNECTIONS = 1500

BRAIN_HZ = 30
MAX_SPEED = 7.0
