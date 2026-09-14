#!/usr/bin/env python3
"""
Fetch neuron data (properties, ROIs, connections, synapses, skeletons)
from a neuPrint server for a given ROOT_NEURON identifier.

Usage:
  python scripts/get_neurons.py --root DNge104 --out data

This script will attempt to import the project's `settings.py` (parent
directory) to pick up `TOKEN`, `SERVER`, and `DATASET`. It also supports
overrides via environment variables or CLI options.
"""

import argparse
import os
import sys
from pathlib import Path

from dotenv import load_dotenv

# from neuprint import NeuronCriteria as NC
from neuprint import Client, fetch_adjacencies, fetch_neurons, fetch_synapses
from neuprint.utils import merge_neuron_properties

load_dotenv()

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROJECT_ROOT))


def get_setting(name, default=None):
    # precedence: env -> settings module -> default
    v = os.environ.get(name)
    if v:
        return v

    return default


def make_client(host, dataset, token):
    # Defer import so script can still be inspected without neuprint installed
    # neuprint Client expects a host without scheme
    host = host.replace("https://", "").replace("http://", "")
    return Client(host, dataset, token)


def main():
    p = argparse.ArgumentParser()
    p.add_argument(
        "--root", required=True, help="ROOT_NEURON name or bodyId (e.g. DNge104)"
    )
    p.add_argument("--out", default="data", help="output directory")
    p.add_argument("--server", default=None, help="neuprint server URL (override)")
    p.add_argument("--dataset", default=None, help="neuprint dataset (override)")
    p.add_argument("--token", default=None, help="neuprint token (override)")
    args = p.parse_args()

    outdir = Path(args.out)
    outdir.mkdir(parents=True, exist_ok=True)

    SERVER = args.server or get_setting("SERVER") or "https://neuprint.janelia.org"
    DATASET = args.dataset or get_setting("DATASET") or "male-cns:v1.0"
    TOKEN = args.token or get_setting("TOKEN") or get_setting("NEUPRINT_TOKEN")

    if not TOKEN:
        raise SystemExit("NEUPRINT token not found in environment or settings")

    print(f"Connecting to neuprint: {SERVER} dataset={DATASET}")
    c = make_client(SERVER, DATASET, TOKEN)

    root = args.root

    print("Fetching neuron properties...")
    neuron_df, roi_counts_df = fetch_neurons(root)

    print("Neuron dataframe shape:", neuron_df)
    print("ROI counts dataframe shape:", roi_counts_df)

    neuron_csv = outdir / f"{root}_neurons.csv"
    roi_csv = outdir / f"{root}_roi_counts.csv"
    neuron_df.to_csv(neuron_csv, index=False)
    roi_counts_df.to_csv(roi_csv, index=False)
    print("Saved", neuron_csv, roi_csv)

    # Fetch connections (adjacencies) FROM this set
    body_ids = list(neuron_df["bodyId"].unique())
    if body_ids:
        print("Fetching adjacencies (connections) from neurons...")
        _, conn_df = fetch_adjacencies(body_ids, None)
        # merge some neuron properties for convenience
        try:
            conn_df = merge_neuron_properties(neuron_df, conn_df, ["type", "instance"])
        except Exception:
            pass
        conn_csv = outdir / f"{root}_connections.csv"
        conn_df.to_csv(conn_csv, index=False)
        print("Saved", conn_csv)

    # Fetch synapses for the matched neurons (may be large)
    try:
        print("Fetching synapses (may take a while)...")
        syn_df = fetch_synapses(root, None)
        syn_csv = outdir / f"{root}_connections.csv"
        syn_df.to_csv(syn_csv, index=False)
        print("Saved", syn_csv)
    except Exception as e:
        print("fetch_synapses failed:", e)

    ids = body_ids
    if ids:
        print("Fetching skeletons for", len(ids), "neurons...")
        for bid in ids:
            try:
                s = c.fetch_skeleton(bid, format="pandas")
                s["bodyId"] = bid
                sk_file = outdir / f"{root}_skeleton_{bid}.csv"
                s.to_csv(sk_file, index=False)
                print("Saved skeleton for", bid)
            except Exception as e:
                print("skeleton fetch failed for", bid, e)

    print("Done.")


if __name__ == "__main__":
    main()
