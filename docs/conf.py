# Configuration file for the Sphinx documentation builder.
# https://www.sphinx-doc.org/en/master/usage/configuration.html

import os
import sys

# Add the source directory so autodoc can find the package.
sys.path.insert(0, os.path.abspath("../src"))

# -- Project information -----------------------------------------------------

project = "fragmented-keys"
copyright = "2024, Keith Brings"
author = "Keith Brings"
release = "0.1.0"

# -- General configuration ---------------------------------------------------

extensions = [
    "sphinx.ext.autodoc",
    "sphinx.ext.intersphinx",
    "sphinx.ext.viewcode",
    "sphinx.ext.napoleon",
]

templates_path = ["_templates"]
exclude_patterns = ["_build", "Thumbs.db", ".DS_Store"]

# -- Options for HTML output -------------------------------------------------

html_theme = "furo"
html_static_path = ["_static"]
html_title = "fragmented-keys"

# -- Extension configuration -------------------------------------------------

autodoc_member_order = "bysource"
autodoc_typehints = "description"

napoleon_google_docstrings = True
napoleon_numpy_docstrings = False

intersphinx_mapping = {
    "python": ("https://docs.python.org/3", None),
    "redis": ("https://redis-py.readthedocs.io/en/stable/", None),
}
