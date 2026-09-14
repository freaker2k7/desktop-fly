class Neuron:
    body_id: int
    name: str
    potential: float
    activity: float

    def __init__(self, body_id, name):
        self.body_id = body_id
        self.name = name
        self.potential = 0.0
        self.activity = 0.0
