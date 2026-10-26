"""Exercise the plotting dependencies without requiring a GUI backend."""

from io import BytesIO
import unittest

import matplotlib

matplotlib.use("Agg")

import matplotlib.pyplot as plt
from matplotlib.font_manager import findfont
from fontTools.ttLib import TTFont
from PIL import Image


class TestPlotting(unittest.TestCase):
    def test_render_formats(self):
        fig, ax = plt.subplots()
        try:
            ax.plot([0, 1, 2], [1, 2, 4], label="Approximation")
            ax.set_title("Approximation smoke test")
            ax.legend()
            for fmt in ("png", "pdf", "svg"):
                with self.subTest(format=fmt):
                    output = BytesIO()
                    fig.savefig(output, format=fmt)
                    self.assertGreater(output.tell(), 0)
                    if fmt == "png":
                        output.seek(0)
                        with Image.open(output) as image:
                            image.load()
                            self.assertGreater(image.width, 0)
                            self.assertGreater(image.height, 0)
        finally:
            plt.close(fig)

    def test_font_loading(self):
        with TTFont(findfont("DejaVu Sans")) as font:
            self.assertIn("head", font)


if __name__ == "__main__":
    unittest.main()
