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

# MAX_NEURONS = 250
# MAX_CONNECTIONS = 1500
MAX_NEURONS = 1e9
MAX_CONNECTIONS = 1e9

BRAIN_HZ = 30
MAX_SPEED = 7.0

# ROI-based weighting: how to modulate a connection's effective weight
# when the connection sits within the same ROI vs across ROIs or when
# ROI information is missing. Tune these between 0.0 and 1.0.
ROI_SAME_BOOST = 1.0
ROI_DIFF_PENALTY = 0.7
ROI_UNKNOWN = 0.85
