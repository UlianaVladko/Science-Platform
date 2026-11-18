from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends, Query
from app.domain.models import Graph
from app.schemas.requests import GraphQuery
from app.application.services import AnalyticsService
from app.api.dependencies import get_analytics_service

router = APIRouter(tags=["Analytics"])


@router.get(
    "/graphs",
    response_model=Graph,
    summary="Get scientific graph",
    description="Small citation, coauthor or concept graph. Root-centered traversal follows edges in both directions up to max_depth. Without root_id, return an overview. max_nodes bounds returned nodes; edges always reference returned nodes.",
    responses=success_response("graph", 200),
)
async def graph(
    service: Annotated[AnalyticsService, Depends(get_analytics_service)],
    query: Annotated[GraphQuery, Query()],
) -> Graph:
    return await service.graph(query)
