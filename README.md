Go OpenGL plot scaffold

This folder contains a minimal Go program that opens an OpenGL window using
`glfw` + `go-gl`. It is a scaffold for the web-plot → OpenGL refactor.

Build (macOS) notes:

- Install GLFW system dependency (Homebrew):

  brew install glfw

- Build using `go build` inside this folder (requires CGO environment):

  cd go
  go mod tidy
  go build -o desktop-fly

- Run:

  ./desktop-fly

Next steps:
- Implement loading neuron/skeleton data (CSV) and draw nodes/edges.
- Implement camera pan so the fly is always centered.
- Replace immediate-mode placeholder with actual GL rendering of graph.
