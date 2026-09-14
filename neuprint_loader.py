import math
import pickle
from pathlib import Path

from brain import Brain
from neuprint import Client, NeuronCriteria, fetch_adjacencies
from settings import DATASET, MAX_CONNECTIONS, MAX_NEURONS, ROOT_NEURON, SERVER, TOKEN


def load_brain():
    cache_dir = Path(".cache")
    cache_dir.mkdir(exist_ok=True)

    # sanitize filename
    safe_name = str(ROOT_NEURON).replace("/", "_")
    cache_file = cache_dir / f"{safe_name}.pkl"

    edges = None
    neurons = None

    if cache_file.exists():
        try:
            print(f"Loading brain from cache: {cache_file}")
            with cache_file.open("rb") as fh:
                neurons, edges = pickle.load(fh)
        except Exception as e:
            print(f"Failed to load cache {cache_file}: {e}")

    if edges is None or neurons is None:
        print("Connecting to neuPrint...")

        client = Client(
            SERVER,
            dataset=DATASET,
            token=TOKEN,
        )

        print(f"Loading connectivity around {ROOT_NEURON}...")

        criteria = NeuronCriteria(type=ROOT_NEURON)

        neurons, edges = fetch_adjacencies(criteria, None, client=client)

        # attempt to cache raw results for faster subsequent runs
        try:
            with cache_file.open("wb") as fh:
                pickle.dump((neurons, edges), fh)
            print(f"Cached neuPrint results to {cache_file}")
        except Exception as e:
            print(f"Failed to write cache {cache_file}: {e}")

    # Normalize returned tables: some versions return (edges, neurons)
    # while others may return (neurons, edges). Detect which is which
    # and ensure `neurons_df` contains neuron rows and `edges_df` contains edges.
    neurons_df = neurons
    edges_df = edges

    if hasattr(neurons_df, "columns") and hasattr(edges_df, "columns"):
        ncols = set(neurons_df.columns)
        ecols = set(edges_df.columns)

        if ("bodyId_pre" in ncols or "bodyId_post" in ncols) and (
            "bodyId" in ecols or "instance" in ecols or "type" in ecols
        ):
            edges_df, neurons_df = neurons_df, edges_df

    if hasattr(neurons_df, "columns") and "bodyId" not in neurons_df.columns:
        for col in neurons_df.columns:
            if "body" in str(col).lower():
                neurons_df = neurons_df.rename(columns={col: "bodyId"})
                break

    brain = Brain()

    if hasattr(neurons_df, "iterrows"):
        for _, row in neurons_df.iterrows():
            # Accept multiple possible column names for the body id.
            if "bodyId" in row:
                body_id = int(row["bodyId"])
            elif "bodyId_pre" in row:
                body_id = int(row["bodyId_pre"])
            elif "bodyId_post" in row:
                body_id = int(row["bodyId_post"])
            else:
                body_id = None
                for col in row.index:
                    if "body" in str(col).lower():
                        try:
                            body_id = int(row[col])
                        except Exception:
                            body_id = None
                        break

            if body_id is None:
                continue

            name = str(
                row.get(
                    "instance",
                    row.get("type", body_id),
                )
            )

            brain.add_neuron(body_id, name)

            if len(brain.neurons) >= MAX_NEURONS:
                break

    if hasattr(edges_df, "iterrows"):
        for _, row in edges_df.iterrows():
            if "bodyId_pre" in row and "bodyId_post" in row:
                src = int(row["bodyId_pre"])
                dst = int(row["bodyId_post"])
            else:
                src = None
                dst = None
                for col in row.index:
                    name = str(col).lower()
                    if "pre" in name and src is None:
                        try:
                            src = int(row[col])
                        except Exception:
                            src = None
                    if "post" in name and dst is None:
                        try:
                            dst = int(row[col])
                        except Exception:
                            dst = None
                if src is None or dst is None:
                    body_cols = [c for c in row.index if "body" in str(c).lower()]
                    if len(body_cols) >= 2:
                        src = int(row[body_cols[0]])
                        dst = int(row[body_cols[1]])
                    else:
                        continue

            if src not in brain.neurons:
                continue

            if dst not in brain.neurons:
                continue

            weight = float(row.get("weight", 1.0))

            # Normalize raw weight into a small bounded value suitable
            # for our simple integrate-and-fire model.
            weight = min(0.25, math.log1p(weight) / 20.0)

            # Extract ROI information if present on the edge row. Many
            # neuPrint exports include an `roi` column indicating the
            # region-of-interest for this connection (e.g. 'GNG'). If
            # present, keep it alongside the edge so the Brain can
            # modulate signal propagation based on region locality.
            roi = None
            if "roi" in row.index and row["roi"] is not None:
                try:
                    roi = str(row["roi"])
                except Exception:
                    roi = None

            # Remember per-neuron ROI if we have it from edges (first
            # seen wins). This helps the Brain compare source/destination
            # regions cheaply at runtime.
            if roi is not None:
                try:
                    if src in brain.neurons:
                        if getattr(brain.neurons[src], "roi", None) is None:
                            brain.neurons[src].roi = roi
                    if dst in brain.neurons:
                        if getattr(brain.neurons[dst], "roi", None) is None:
                            brain.neurons[dst].roi = roi
                except Exception:
                    pass

            # Store ROI with the edge as a 4-tuple. The Brain accepts both
            # (src,dst,weight) and (src,dst,weight,roi) formats for
            # backward compatibility.
            brain.edges.append((src, dst, weight, roi if roi is not None else None))

            if len(brain.edges) >= MAX_CONNECTIONS:
                break

    if not brain.neurons:
        raise RuntimeError("neuPrint returned no neurons.")

    print(f"Loaded {len(brain.neurons)} neurons and {len(brain.edges)} connections.")

    return brain
