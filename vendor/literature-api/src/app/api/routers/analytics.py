from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends, Query
from app.domain.models import AggregateResult, CorpusSummary, Ranking
from app.schemas.requests import AggregateQuery, RankingQuery
from app.application.services import AnalyticsService
from app.api.dependencies import get_analytics_service

router = APIRouter(tags=["Analytics"])
Service = Annotated[AnalyticsService, Depends(get_analytics_service)]


@router.get(
    "/analytics/summary",
    response_model=CorpusSummary,
    summary="Get corpus summary",
    description="Return fixture counters; this_period refers to the fixed demonstration year 2026.",
    responses=success_response("summary", 200),
)
async def summary(service: Service) -> CorpusSummary:
    return await service.summary()


@router.get(
    "/analytics/aggregates",
    response_model=AggregateResult,
    summary="Get analytical aggregates",
    description="One or two comma-separated dimensions. Only documented metric/dimension combinations are supported. Two dimensions form heatmap data. period is a demonstration publication year filter, not an arbitrary date expression.",
    responses=success_response("aggregates", 200),
)
async def aggregates(
    service: Service, query: Annotated[AggregateQuery, Query()]
) -> AggregateResult:
    return await service.aggregates(query)


@router.get(
    "/analytics/top",
    response_model=Ranking,
    summary="Get rankings",
    description="Publication citations/views/downloads, author citations/publications, keyword/concept frequency. Other combinations fail validation.",
    responses=success_response("top", 200),
)
async def top(service: Service, query: Annotated[RankingQuery, Query()]) -> Ranking:
    return await service.top(query)
