# Desktop Fly

[![GitHub Build](https://github.com/freaker2k7/desktop-fly/actions/workflows/ci.yml/badge.svg)](https://github.com/freaker2k7/desktop-fly/actions)

A minimal desktop application for visualizing a fly flying over your desktop and its neuron data using OpenGL.
The fly "thinks" in terms of its neuron activity as we stream desktop screenshots and mouse movements into the neural network.
The neuron data is taken from [male-cns:v1.0](https://neuprint.janelia.org/) for a male Drosophila (fruit fly) central nervous system 🪰

I didn't use the whole dataset, only a few neuron types:

- DNge104 - [Descending neuron](https://en.wikipedia.org/wiki/Descending_neuron), for the motor control
- KC - [Kenyon cell](https://en.wikipedia.org/wiki/Kenyon_cell), for the [mushroom body](https://en.wikipedia.org/wiki/Mushroom_bodies) and learning processes
- HBeyelet - [Horizontal system B eyelet](https://www.researchgate.net/publication/379948772_Connectome-driven_neural_inventory_of_a_complete_visual_system), for the visual system
- TTMn - [Tergotrochanteral motor neuron](https://pubmed.ncbi.nlm.nih.gov/11060245/), for the leg movement

## Features

- Visualize a fly flying over your desktop using OpenGL.
- Visualize fly neuron data using OpenGL.
- Minimal and lightweight desktop application.
- Cross-platform support (Windows, macOS, Linux).

## Installation

1. Download the latest release from the [releases page](https://github.com/freaker2k7/desktop-fly/releases) for your operating system.
2. Extract the downloaded archive to a desired location.
3. Navigate to the extracted directory and run the application 🤝

*NOTE: On macOS, you may need to allow the application to run from the Security & Privacy settings if it is blocked by Gatekeeper.*

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

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/freaker2k7/desktop-fly?tab=MIT-1-ov-file) file for details.
