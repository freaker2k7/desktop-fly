import pyglet
from fly_window import FlyWindow
from neuprint_loader import load_brain


def main():
    brain = load_brain()

    window = FlyWindow(brain)

    pyglet.app.run()


if __name__ == "__main__":
    main()
