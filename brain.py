from neuron import Neuron
from settings import ROI_DIFF_PENALTY, ROI_SAME_BOOST, ROI_UNKNOWN


class Brain:
    """Simple spiking-like network used to control the fly.

    - `self.neurons` maps body_id -> `Neuron` objects.
    - `self.edges` is a list of (src, dst, weight) tuples describing
      directed connections between neurons.
    """

    def __init__(self):
        # store neurons by id for O(1) lookup when applying edges
        self.neurons = {}
        # list of (src_id, dst_id, weight)
        self.edges = []

    def add_neuron(self, body_id, name):
        # only add a neuron once (avoid overwriting existing state)
        if body_id not in self.neurons:
            self.neurons[body_id] = Neuron(body_id, name)

    def step(self, sensory: dict[str, float]):
        """Advance the brain by one timestep using the provided sensory dict.

        The function returns a (turn, thrust) tuple where `turn` is in
        [-1.0, 1.0] and `thrust` is in [0.0, 1.0].

        Expected `sensory` keys: 'left', 'right', 'mouse', 'brightness'.
        """

        # Copy current neurons into a list for stable iteration order.
        neurons = list(self.neurons.values())

        # If there are no neurons loaded, return neutral motor commands.
        if not neurons:
            return 0.0, 0.0

        # Read sensory channels into local variables for clarity.
        left = sensory["left"]
        right = sensory["right"]
        mouse = sensory["mouse"]
        brightness = sensory["brightness"]

        # Desktop -> neural input mapping:
        # Only the first 8 neurons receive direct sensory input. This keeps
        # the amount of external drive bounded and simulates a small set
        # of sensory neurons projecting into the network.
        # The multipliers are empirically chosen gains:
        #  - 0.35 for left/right: moderate influence from lateral signals
        #  - 0.5  for mouse: stronger attractor towards the mouse cursor
        #  - 0.15 for brightness: weaker modulation from ambient brightness
        # The use of `i % 4` interleaves the four channels across those 8
        # neurons so that the input is distributed rather than concentrated.
        for i, neuron in enumerate(neurons[:8]):
            if i % 4 == 0:
                neuron.potential += left * 0.35
            elif i % 4 == 1:
                neuron.potential += right * 0.35
            elif i % 4 == 2:
                neuron.potential += mouse * 0.5
            else:
                neuron.potential += brightness * 0.15

        # `spikes` collects ids of neurons that crossed threshold this step.
        spikes = set()

        # Leak and spike generation loop:
        # - Multiply potential by 0.92 each step to model exponential leak
        #   (0.92 means ~8% decay per timestep).
        # - If potential >= 1.0, consider that a spike: reset potential to 0
        #   and set activity to 1.0 (a binary spike-like signal). Threshold
        #   of 1.0 is used as a convenient normalized firing threshold.
        # - Otherwise, decay the recent activity by 0.75 to model
        #   short-lived activity traces (75% retained each step).
        for neuron in neurons:
            neuron.potential *= 0.92

            if neuron.potential >= 1.0:
                # neuron fired: reset potential and mark activity
                neuron.potential = 0.0
                neuron.activity = 1.0
                spikes.add(neuron.body_id)
            else:
                # decay the activity trace if no spike occurred
                neuron.activity *= 0.75

        # Propagate spikes along edges: if a source neuron spiked this step
        # add `weight` to the target neuron's potential. Edges are applied
        # synchronously (all spikes from this timestep are delivered once).
        for edge in self.edges:
            # Support both (src,dst,weight) and (src,dst,weight,roi)
            if len(edge) == 3:
                src, dst, weight = edge
                edge_roi = None
            else:
                src, dst, weight, edge_roi = edge

            if src in spikes:
                target = self.neurons.get(dst)
                if target:
                    # Compute a small multiplier based on ROI locality. If
                    # the connection sits within the same ROI we keep the
                    # full weight; cross-ROI connections are slightly
                    # penalized to enforce locality.
                    try:
                        src_roi = getattr(self.neurons.get(src), "roi", None)
                        dst_roi = getattr(target, "roi", None)
                    except Exception:
                        src_roi = None
                        dst_roi = None

                    if edge_roi:
                        # If the edge itself reports an ROI, prefer that
                        # information when deciding locality.
                        same_roi = (src_roi is not None and src_roi == edge_roi) or (
                            dst_roi is not None and dst_roi == edge_roi
                        )
                    else:
                        same_roi = src_roi is not None and src_roi == dst_roi

                    if same_roi:
                        mult = ROI_SAME_BOOST
                    elif edge_roi is None and src_roi is None and dst_roi is None:
                        mult = ROI_UNKNOWN
                    else:
                        mult = ROI_DIFF_PENALTY

                    target.potential += weight * mult

        # Convert neuron activities into motor commands.
        # `turn` and `thrust` are accumulated by scanning neuron activities.
        # The modulo distribution assigns every third neuron to one of the
        # three channels: turn-left, turn-right, and thrust. This is a
        # simple way to create non-linear mixing without learned readouts.
        turn = 0.0
        thrust = 0.0

        for i, neuron in enumerate(neurons):
            if i % 3 == 0:
                # neurons at indices 0,3,6,... bias turning one direction
                turn -= neuron.activity
            elif i % 3 == 1:
                # neurons at indices 1,4,7,... bias turning the opposite way
                turn += neuron.activity
            else:
                # neurons at indices 2,5,8,... contribute to forward thrust
                thrust += neuron.activity

        # Scale the outputs by the network size to keep commands in a reasonable
        # range. Using `len(neurons) // 8` groups neurons in blocks of 8; the
        # `max(1, ...)` prevents division by zero and ensures small networks
        # still produce meaningful commands.
        scale = max(1, len(neurons) // 8)

        # Normalize and clamp outputs to expected ranges:
        # - turn: [-1.0, 1.0]
        # - thrust: [0.0, 1.0]
        return (
            max(-1.0, min(1.0, turn / scale)),
            max(0.0, min(1.0, thrust / scale)),
        )
