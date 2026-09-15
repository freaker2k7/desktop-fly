# Desktop Fly

[![GitHub Build](https://github.com/freaker2k7/desktop-fly/actions/workflows/ci.yml/badge.svg)](https://github.com/freaker2k7/desktop-fly/actions)

A minimal desktop application for visualizing fly neuron data using OpenGL.

## Features

- Visualize fly neuron data in 3D using OpenGL.
- Load neuron and skeleton data from CSV files.
- Interactive camera controls for exploring the 3D scene.
- Support for multiple neuron datasets and easy switching between them.
- Export visualizations as images or 3D models.

## Installation

1. Download the latest release from the [releases page](https://github.com/freaker2k7/desktop-fly/releases) for your operating system.
2. Extract the downloaded archive to a desired location.
3. Navigate to the extracted directory and run the application.

NOTE: On macOS, you may need to allow the application to run from the Security & Privacy settings if it is blocked by Gatekeeper.

## Development

### Prerequisites

- Go (https://golang.org/dl/)
- Python (https://www.python.org/downloads/)
- OpenGL (https://www.opengl.org/) [Usually comes pre-installed on most systems]

### Installation

```bash
git clone https://github.com/freaker2k7/desktop-fly.git
cd desktop-fly
pip install -r requirements.txt
```

### Create Dataset

```bash
python scripts/get_neurons.py --root DNge104
python scripts/get_neurons.py --root KC
python scripts/get_neurons.py --root HBeyelet
python scripts/get_neurons.py --root TTMn
```

### Build and Run

```bash
go build
go test
./desktop-fly
```