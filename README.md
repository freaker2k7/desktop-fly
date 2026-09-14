# Desktop Fly

A minimal desktop application for visualizing fly neuron data using OpenGL.

## Features

- Visualize fly neuron data in 3D using OpenGL.
- Load neuron and skeleton data from CSV files.
- Interactive camera controls for exploring the 3D scene.
- Support for multiple neuron datasets and easy switching between them.
- Export visualizations as images or 3D models.

## Installation

```bash
git clone https://github.com/yourusername/desktop-fly.git
cd desktop-fly
pip install -r requirements.txt
```

## Create Dataset

```bash
python scripts/get_neurons.py --root KC
python scripts/get_neurons.py --root HBeyelet
python scripts/get_neurons.py --root TTMn
```

# Build and Run

```bash
go build
./desktop-fly
```