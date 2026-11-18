"""FastAPI application factory; every instance owns independent mock state."""

from fastapi import FastAPI, APIRouter
from app.api.dependencies import build_services
from app.api.exception_handlers import RequestIDMiddleware, register_handlers
from app.api.routers import (
    categories,
    search,
    publications,
    authors,
    analytics,
    graphs,
    jobs,
    clustering,
)
from app.schemas.errors import ERROR_RESPONSES


def create_app() -> FastAPI:
    app = FastAPI(
        title="Scientific Literature Analytics API",
        version="1.0.0",
        description="External API skeleton with in-memory mock adapters. No databases, search engines, NLP or background workers are required.",
        openapi_tags=[
            {"name": name, "description": description}
            for name, description in [
                ("Classification", "Discipline navigation"),
                ("Search", "Discovery and suggestions"),
                ("Publications", "Metadata, files and bibliography"),
                ("Authors", "Researcher profiles"),
                ("Analytics", "Aggregates, rankings and graphs"),
                ("Analysis", "Jobs and clustering"),
            ]
        ],
    )
    app.state.services = build_services()
    app.add_middleware(RequestIDMiddleware)
    register_handlers(app)
    public = APIRouter(prefix="/api/v1", responses=ERROR_RESPONSES)
    for module in (
        categories,
        search,
        publications,
        authors,
        analytics,
        graphs,
        jobs,
        clustering,
    ):
        public.include_router(module.router)
    app.include_router(public)
    return app


app = create_app()
