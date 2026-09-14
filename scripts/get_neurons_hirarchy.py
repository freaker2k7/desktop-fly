import os
from pathlib import Path

import pandas as pd
from dotenv import load_dotenv
from neuprint import Client, fetch_neurons

# ============================================================
# Configuration
# ============================================================

load_dotenv()

TOKEN = os.getenv("NEUPRINT_TOKEN")

if not TOKEN:
    raise RuntimeError("NEUPRINT_TOKEN was not found. " "Add it to your .env file.")

DATASET = "male-cns:v1.0"
OUTPUT_DIR = Path("data")

OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

CELL_TYPES_FILE = OUTPUT_DIR / "male_cns_cell_types.csv"
SUMMARY_FILE = OUTPUT_DIR / "male_cns_summary.txt"


# ============================================================
# Connect to neuPrint
# ============================================================

client = Client("https://neuprint.janelia.org", dataset=DATASET, token=TOKEN)


# ============================================================
# Fetch neuron metadata
#
# This retrieves neuron metadata only.
# It does NOT download skeletons or synapse coordinates.
# ============================================================

neurons, synapse_distribution = fetch_neurons(client=client)


# ============================================================
# Helper functions
# ============================================================


def unique_values(series):
    """
    Convert a pandas Series into a sorted, unique string.

    Handles normal values as well as lists/tuples/sets.
    """

    values = []

    for value in series.dropna():

        if isinstance(value, (list, tuple, set)):
            values.extend(value)
        else:
            values.append(value)

    return "; ".join(sorted(set(str(value) for value in values)))


def get_column_values(group, column):
    """
    Return unique values from a column if it exists.
    """

    if column not in group.columns:
        return ""

    return unique_values(group[column])


def infer_function(row):
    """
    Conservative heuristic for assigning a broad functional category.

    IMPORTANT:
    This is an inference from metadata, not an official biological
    annotation. If the available metadata is insufficient, return
    'Unknown'.
    """

    text = " ".join(
        str(row.get(column, ""))
        for column in [
            "cell_type",
            "class",
            "superclass",
            "hemilineage",
            "modality",
            "entry_nerve",
            "exit_nerve",
            "description",
        ]
    ).lower()

    # --------------------------------------------------------
    # Sensory
    # --------------------------------------------------------

    sensory_keywords = [
        "visual",
        "optic",
        "photoreceptor",
        "olfactory",
        "gustatory",
        "auditory",
        "mechanosensory",
        "thermosensory",
        "proprio",
    ]

    if any(keyword in text for keyword in sensory_keywords):
        return "Sensory / sensory processing"

    # --------------------------------------------------------
    # Motor
    # --------------------------------------------------------

    motor_keywords = [
        "motor neuron",
        "motor",
        "motoneuron",
    ]

    if any(keyword in text for keyword in motor_keywords):
        return "Motor / movement"

    # --------------------------------------------------------
    # Descending
    # --------------------------------------------------------

    descending_keywords = [
        "descending",
    ]

    if any(keyword in text for keyword in descending_keywords):
        return "Descending pathway: brain → VNC"

    # Common DN naming convention
    cell_type = str(row.get("cell_type", "")).lower()

    if cell_type.startswith("dn"):
        return "Likely descending neuron"

    # --------------------------------------------------------
    # Ascending
    # --------------------------------------------------------

    ascending_keywords = [
        "ascending",
    ]

    if any(keyword in text for keyword in ascending_keywords):
        return "Ascending pathway: VNC → brain"

    # --------------------------------------------------------
    # Neuromodulatory / neurotransmitter-defined
    # --------------------------------------------------------

    neuromodulatory_keywords = [
        "dopaminergic",
        "dopamine",
        "serotonergic",
        "serotonin",
        "octopaminergic",
        "octopamine",
        "gabaergic",
        "cholinergic",
    ]

    if any(keyword in text for keyword in neuromodulatory_keywords):
        return "Neuromodulatory / neurotransmitter-defined"

    # --------------------------------------------------------
    # Central processing
    # --------------------------------------------------------

    central_keywords = [
        "interneuron",
        "central",
        "local",
        "projection",
    ]

    if any(keyword in text for keyword in central_keywords):
        return "Central processing / interneuron"

    # --------------------------------------------------------
    # Unknown
    # --------------------------------------------------------

    return "Unknown"


# ============================================================
# Prepare cell-type grouping
# ============================================================

# Prefer the annotated cell type.
# If type is missing, fall back to instance so that neurons
# are not silently discarded.

if "type" in neurons.columns:
    neurons["cell_type"] = neurons["type"]

    if "instance" in neurons.columns:
        neurons["cell_type"] = neurons["cell_type"].fillna(neurons["instance"])

else:
    neurons["cell_type"] = neurons["instance"]


# ============================================================
# Build one record per cell type
# ============================================================

records = []

for cell_type, group in neurons.groupby("cell_type", dropna=False):

    record = {
        # ----------------------------------------------------
        # Identity
        # ----------------------------------------------------
        "cell_type": cell_type,
        "n_neurons": len(group),
        "instances": get_column_values(
            group,
            "instance",
        ),
        "sides": get_column_values(
            group,
            "side",
        ),
        # ----------------------------------------------------
        # Classification
        # ----------------------------------------------------
        "class": get_column_values(
            group,
            "class",
        ),
        "superclass": get_column_values(
            group,
            "superclass",
        ),
        "hemilineage": get_column_values(
            group,
            "hemilineage",
        ),
        "modality": get_column_values(
            group,
            "modality",
        ),
        # ----------------------------------------------------
        # Nerve information
        # ----------------------------------------------------
        "entry_nerve": get_column_values(
            group,
            "entryNerve",
        ),
        "exit_nerve": get_column_values(
            group,
            "exitNerve",
        ),
        # ----------------------------------------------------
        # Text annotations
        # ----------------------------------------------------
        "description": get_column_values(
            group,
            "description",
        ),
        "synonyms": get_column_values(
            group,
            "synonyms",
        ),
    }

    # --------------------------------------------------------
    # Functional inference
    # --------------------------------------------------------

    record["likely_function"] = infer_function(record)

    records.append(record)


# ============================================================
# Create DataFrame
# ============================================================

cell_types = pd.DataFrame(records)


# ============================================================
# Sort
# ============================================================

sort_columns = [
    column
    for column in [
        "class",
        "superclass",
        "cell_type",
    ]
    if column in cell_types.columns
]

cell_types = cell_types.sort_values(
    sort_columns,
    na_position="last",
).reset_index(drop=True)


# ============================================================
# Save cell-type catalogue
# ============================================================

cell_types.to_csv(
    CELL_TYPES_FILE,
    index=False,
)


# ============================================================
# Create summary
# ============================================================

summary = [
    "MaleCNS Cell-Type Catalogue",
    "=" * 60,
    "",
    f"Dataset: {DATASET}",
    "",
    f"Neuron instances: {len(neurons):,}",
    f"Cell types: {len(cell_types):,}",
    "",
    "Output:",
    f"  {CELL_TYPES_FILE}",
    "",
    "Columns:",
]

summary.extend(f"  - {column}" for column in cell_types.columns)


# ============================================================
# Add class statistics
# ============================================================

if "class" in cell_types.columns:

    summary.extend(
        [
            "",
            "Cell types by class:",
        ]
    )

    class_counts = cell_types["class"].replace("", pd.NA).dropna().value_counts()

    for class_name, count in class_counts.items():
        summary.append(f"  {class_name}: {count:,}")


# ============================================================
# Add functional statistics
# ============================================================

if "likely_function" in cell_types.columns:

    summary.extend(
        [
            "",
            "Cell types by inferred function:",
        ]
    )

    function_counts = cell_types["likely_function"].value_counts()

    for function, count in function_counts.items():
        summary.append(f"  {function}: {count:,}")


# ============================================================
# Save summary
# ============================================================

SUMMARY_FILE.write_text(
    "\n".join(summary),
    encoding="utf-8",
)
