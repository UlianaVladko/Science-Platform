from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends, Path
from app.domain.models import AnalysisJob
from app.schemas.requests import JobRequest
from app.application.services import AnalysisJobService
from app.api.dependencies import get_analysisjob_service

router = APIRouter(tags=["Analysis"])
Service = Annotated[AnalysisJobService, Depends(get_analysisjob_service)]


@router.post(
    "/analysis/jobs",
    status_code=202,
    response_model=AnalysisJob,
    summary="Submit analytical job",
    description="Accept a typed processing, term extraction or clustering request. New mock jobs remain pending; clustering also creates a pending run. No computations or background workers are started.",
    openapi_extra={
        "requestBody": {
            "content": {
                "application/json": {
                    "example": {
                        "operation": "clustering",
                        "publication_ids": ["pub-1", "pub-2"],
                        "parameters": {"algorithm": "kmeans", "n_clusters": 2},
                    }
                }
            }
        }
    },
    responses=success_response("start_job", 202),
)
async def start_job(command: JobRequest, service: Service) -> AnalysisJob:
    return await service.start(command)


@router.get(
    "/jobs/{job_id}",
    response_model=AnalysisJob,
    summary="Get job state",
    description="Read pending/running/completed/failed demo jobs (job-pending, job-running, job-completed, job-failed), or a newly accepted task.",
    responses=success_response("job", 200),
)
async def job(
    job_id: Annotated[str, Path(min_length=1, examples=["job-completed"])],
    service: Service,
) -> AnalysisJob:
    return await service.get(job_id)
