import math
import threading
import time

import pyglet
from fly import Fly
from pyglet import shapes
from settings import BRAIN_HZ


class FlyWindow(pyglet.window.Window):

    def __init__(self, brain):
        display = pyglet.display.get_display()
        screen = display.get_default_screen()

        self.screen_width = screen.width
        self.screen_height = screen.height

        super().__init__(
            width=self.screen_width,
            height=self.screen_height,
            caption="Desktop Fly",
            style=pyglet.window.Window.WINDOW_STYLE_BORDERLESS,
            vsync=True,
            resizable=False,
        )

        self.set_location(0, 0)

        self.brain = brain

        self.fly = Fly(
            self.screen_width,
            self.screen_height,
        )

        self.mouse_x = self.screen_width / 2
        self.mouse_y = self.screen_height / 2

        self.running = True

        # Try to make the window behave like a desktop overlay.
        try:
            self.set_always_on_top(True)
        except Exception:
            pass

        # Transparent background.
        try:
            self.set_opacity(0)
        except Exception:
            pass

        self._create_fly_graphics()

        self.brain_thread = threading.Thread(
            target=self._brain_loop,
            daemon=True,
        )

        self.brain_thread.start()

        pyglet.clock.schedule_interval(
            self.update,
            1.0 / 60.0,
        )

    def _create_fly_graphics(self):
        self.body = shapes.Ellipse(
            0,
            0,
            13,
            7,
        )

        self.head = shapes.Circle(
            0,
            0,
            7,
        )

        self.wing1 = shapes.Ellipse(
            0,
            0,
            22,
            6,
        )

        self.wing2 = shapes.Ellipse(
            0,
            0,
            22,
            6,
        )

        self.eye1 = shapes.Circle(
            0,
            0,
            2,
        )

        self.eye2 = shapes.Circle(
            0,
            0,
            2,
        )

        self.batch = pyglet.graphics.Batch()

        # Dark fly.
        self.body.color = (25, 25, 25)
        self.head.color = (45, 45, 45)

        # Semi-transparent wings.
        self.wing1.color = (180, 210, 240)
        self.wing2.color = (180, 210, 240)

        self.eye1.color = (220, 40, 40)
        self.eye2.color = (220, 40, 40)

    def _brain_loop(self):
        while self.running:
            sensors = self.fly.sensors(
                self.mouse_x,
                self.mouse_y,
            )

            turn, thrust = self.brain.step(sensors)

            self.fly.turn = turn
            self.fly.thrust = thrust

            time.sleep(1.0 / BRAIN_HZ)

    def on_mouse_motion(self, x, y, dx, dy):
        self.mouse_x = x
        self.mouse_y = y

    def on_mouse_press(self, x, y, button, modifiers):
        # Right click quits.
        if button == pyglet.window.mouse.RIGHT:
            self.close()

    def on_key_press(self, symbol, modifiers):
        if symbol == pyglet.window.key.ESCAPE:
            self.close()

    def _position_graphics(self):
        x = self.fly.x
        y = self.fly.y

        angle = self.fly.angle

        dx = math.cos(angle)
        dy = math.sin(angle)

        px = -dy
        py = dx

        # Body.
        self.body.x = x - dx * 5
        self.body.y = y - dy * 5

        self.body.rotation = -math.degrees(angle)

        # Head.
        self.head.x = x + dx * 11
        self.head.y = y + dy * 11

        # Wings.
        self.wing1.x = x + px * 10
        self.wing1.y = y + py * 10
        self.wing1.rotation = -math.degrees(angle) + 20

        self.wing2.x = x - px * 10
        self.wing2.y = y - py * 10
        self.wing2.rotation = -math.degrees(angle) - 20

        # Eyes.
        self.eye1.x = x + dx * 14 + px * 4
        self.eye1.y = y + dy * 14 + py * 4

        self.eye2.x = x + dx * 14 - px * 4
        self.eye2.y = y + dy * 14 - py * 4

    def on_draw(self):
        self.clear()

        self._position_graphics()

        self.wing1.draw()
        self.wing2.draw()
        self.body.draw()
        self.head.draw()
        self.eye1.draw()
        self.eye2.draw()

    def update(self, dt):
        if not self.running:
            return

        self.fly.update(min(dt, 0.05))

    def close(self):
        self.running = False
        super().close()
