from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends, Query, Path
from app.domain.models import ClusteringRun, ClusteringRunDetail, Page
from app.schemas.requests import Pagination
from app.application.services import AnalysisJobService
from app.api.dependencies import get_analysisjob_service

router = APIRouter(tags=["Analysis"])
Service = Annotated[AnalysisJobService, Depends(get_analysisjob_service)]


@router.get(
    "/analysis/clustering/runs",
    response_model=Page[ClusteringRun],
    summary="List clustering runs",
    description="Paginated fixture run summaries without cluster membership.",
    responses=success_response("runs", 200),
)
async def runs(
    service: Service, pagination: Annotated[Pagination, Query()]
) -> Page[ClusteringRun]:
    return await service.runs(pagination)


@router.get(
    "/analysis/clustering/runs/{run_id}",
    response_model=ClusteringRunDetail,
    summary="Get clustering result",
    description="Typed parameters, metrics and cluster membership. run-1 is a completed deterministic demonstration.",
    responses=success_response("run", 200),
)
async def run(
    run_id: Annotated[str, Path(min_length=1, examples=["run-1"])], service: Service
) -> ClusteringRunDetail:
    return await service.run(run_id)
