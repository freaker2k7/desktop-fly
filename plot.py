import math
import threading
import webbrowser
from typing import Iterable

import hvplot.pandas  # registers hvplot accessor
import pandas as pd

try:
    import holoviews as hv
    import panel as pn

    hv.extension("bokeh")
    pn.extension()
    _PANEL_AVAILABLE = True
except Exception:
    _PANEL_AVAILABLE = False


class Plot:
    """Network plot using hvplot; prefers a live Panel server when available.

    Behaviour:
      - If `panel` is installed, starts a background Panel server and
        serves a live-updating view of the network.
      - Otherwise falls back to saving static HTML on `show()` / `update()`.
    """

    def __init__(self, brain, node_size=6, triggered_size=16):
        self.brain = brain
        self.node_size = node_size
        self.triggered_size = triggered_size

        # deterministic layout on unit sphere projected to 2D (x,y)
        self.node_ids = list(self.brain.neurons.keys())
        self.pos = self._init_positions(len(self.node_ids))

        # Panel state
        self._server_thread = None
        self._pane = None

        if _PANEL_AVAILABLE:
            # create initial pane and serve it
            self._pane = pn.pane.HoloViews(self._build_hv(), sizing_mode="stretch_both")
            self._start_panel_server()

    def _init_positions(self, n):
        pos = {}
        if n == 0:
            return pos
        phi = math.pi * (3.0 - math.sqrt(5.0))
        for i, nid in enumerate(self.node_ids):
            y = 1 - (i / float(max(1, n - 1))) * 2
            radius = math.sqrt(max(0.0, 1 - y * y))
            theta = phi * i
            x = math.cos(theta) * radius
            z = math.sin(theta) * radius
            pos[nid] = (x, y, z)
        return pos

    def _build_hv(self, triggered_ids: Iterable[int] | None = None):
        # build node dataframe
        rows = []
        triggered = set(triggered_ids) if triggered_ids is not None else None
        for nid in self.node_ids:
            x, y, _ = self.pos[nid]
            neuron = self.brain.neurons[nid]
            activity = getattr(neuron, "activity", 0.0)
            if triggered is None:
                color = activity
                size = self.node_size + activity * self.node_size
            else:
                color = 1.0 if nid in triggered else 0.2
                size = self.triggered_size if nid in triggered else self.node_size
            rows.append(
                {
                    "nid": nid,
                    "name": neuron.name,
                    "x": x,
                    "y": y,
                    "color": color,
                    "size": size,
                }
            )

        nodes = pd.DataFrame(rows)

        # build edges as segments
        seg_rows = []
        for e in self.brain.edges:
            if len(e) >= 2:
                src, dst = e[0], e[1]
            else:
                continue
            if src not in self.pos or dst not in self.pos:
                continue
            x0, y0, _ = self.pos[src]
            x1, y1, _ = self.pos[dst]
            seg_rows.append({"x0": x0, "y0": y0, "x1": x1, "y1": y1})

        edges = pd.DataFrame(seg_rows)

        # use holoviews elements for overlay
        try:
            import holoviews as hv

            points = nodes.hvplot.scatter(
                x="x",
                y="y",
                size="size",
                color="color",
                hover_cols=["name"],
                cmap="Viridis",
            )
            segments = hv.Segments(edges)
            return segments * points
        except Exception:
            # fallback to hvplot only
            p = nodes.hvplot.scatter(
                x="x",
                y="y",
                size="size",
                color="color",
                hover_cols=["name"],
                cmap="Viridis",
            )
            return p

    def _start_panel_server(self):
        def serve():
            try:
                pn.serve(self._pane, show=True, start=True)
            except Exception:
                # best-effort: if pn.serve fails, ignore
                pass

        t = threading.Thread(target=serve, daemon=True)
        t.start()
        self._server_thread = t

    def update(self, triggered_ids: Iterable[int] | None = None):
        # update the pane if panel is available
        if _PANEL_AVAILABLE and self._pane is not None:
            try:
                self._pane.object = self._build_hv(triggered_ids)
                return
            except Exception:
                pass

        # fallback: regenerate static html view
        self.show(
            filename="brain_plot.html", auto_open=False, triggered_ids=triggered_ids
        )

    def show(
        self,
        filename: str | None = None,
        auto_open: bool = True,
        triggered_ids: Iterable[int] | None = None,
    ):
        """Render the view. If Panel is running this will have already been shown.

        If Panel is not available, writes a static HTML file.
        """
        if _PANEL_AVAILABLE:
            # panel server already opened a browser tab; just update object
            if self._pane is not None:
                self._pane.object = self._build_hv(triggered_ids)
            return

        # static fallback using holoviews hv.save
        try:
            import holoviews as hv

            hv_obj = self._build_hv(triggered_ids)
            if filename is None:
                # render inline (may open js-backed viewer)
                hv.output(hv_obj)
                return

            hv.save(hv_obj, filename)
            if auto_open:
                webbrowser.open(filename)
        except Exception:
            # last-resort: do nothing
            pass
