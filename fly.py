import math
import random

from settings import MAX_SPEED


class Fly:
    """A class representing a fly in the simulation."""

    width: float
    height: float
    x: float
    y: float
    vx: float
    vy: float
    angle: float
    turn: float
    thrust: float
    wander: float

    def __init__(self, width: float, height: float):
        self.width = width
        self.height = height

        self.x = width * 0.5
        self.y = height * 0.5

        self.vx = random.uniform(-1, 1)
        self.vy = random.uniform(-1, 1)

        self.angle = random.random() * math.tau

        self.turn = 0.0
        self.thrust = 0.0

        self.wander = 0.0

    def sensors(self, mouse_x, mouse_y):
        dx = mouse_x - self.x
        dy = mouse_y - self.y

        distance = math.hypot(dx, dy)

        mouse_signal = max(
            0.0,
            1.0 - distance / 300.0,
        )

        # Mouse position relative to fly heading.
        c = math.cos(-self.angle)
        s = math.sin(-self.angle)

        local_x = dx * c - dy * s

        left = max(
            0.0,
            -local_x / 400.0,
        )

        right = max(
            0.0,
            local_x / 400.0,
        )

        brightness = (abs(math.sin(self.x * 0.01)) + abs(math.cos(self.y * 0.01))) * 0.5

        return {
            "mouse": mouse_signal,
            "left": min(1.0, left),
            "right": min(1.0, right),
            "brightness": brightness,
        }

    def update(self, dt):
        self.wander += random.uniform(-0.15, 0.15)

        self.angle += self.turn * 0.08 + self.wander * 0.002

        acceleration = self.thrust * 0.35 + 0.04

        self.vx += math.cos(self.angle) * acceleration
        self.vy += math.sin(self.angle) * acceleration

        self.vx *= 0.985
        self.vy *= 0.985

        speed = math.hypot(self.vx, self.vy)

        if speed > MAX_SPEED:
            factor = MAX_SPEED / speed
            self.vx *= factor
            self.vy *= factor

        self.x += self.vx
        self.y += self.vy

        margin = 20

        if self.x < margin:
            self.x = margin
            self.vx = abs(self.vx)

        elif self.x > self.width - margin:
            self.x = self.width - margin
            self.vx = -abs(self.vx)

        if self.y < margin:
            self.y = margin
            self.vy = abs(self.vy)

        elif self.y > self.height - margin:
            self.y = self.height - margin
            self.vy = -abs(self.vy)
