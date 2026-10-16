"""ASGI entry point: `uvicorn harbormaster.main:app`."""

from .web.app import create_app

app = create_app()
