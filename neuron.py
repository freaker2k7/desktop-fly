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
        # Region-of-interest label (if known). Kept as a small string
        # to allow region-aware routing/weighting in the Brain.
        self.roi = None
