class Neuron:
    body_id: int
    name: str
    potential: float
    activity: float
    roi: str

    def __init__(self, body_id, name):
        self.body_id = body_id
        self.name = name
        # A potential value representing the neuron's readiness to fire.
        self.potential = 0.0
        # A value representing the neuron's recent activity.
        self.activity = 0.0
        # Region-of-interest label (if known). Kept as a small string
        # to allow region-aware routing/weighting in the Brain.
        self.roi = None
