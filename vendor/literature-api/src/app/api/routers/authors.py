from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends, Path
from app.domain.models import Author

from app.application.services import AuthorService
from app.api.dependencies import get_author_service

router = APIRouter(tags=["Authors"])


@router.get(
    "/authors/{id}",
    response_model=Author,
    summary="Get researcher profile",
    description="Researcher affiliations, disciplines and activity metrics. Obtain publications using POST /search with filters.author.",
    responses=success_response("author", 200),
)
async def author(
    id: Annotated[str, Path(min_length=1, examples=["author-1"])],
    service: Annotated[AuthorService, Depends(get_author_service)],
) -> Author:
    return await service.get(id)
