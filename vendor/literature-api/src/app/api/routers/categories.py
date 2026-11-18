from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends
from app.domain.models import Category

from app.application.services import CategoryService
from app.api.dependencies import get_category_service

router = APIRouter(tags=["Classification"])


@router.get(
    "/categories",
    response_model=list[Category],
    summary="Get discipline tree",
    description="Return root categories, or children of parent_id, with nested children and optional subtree publication counts.",
    responses=success_response("categories", 200),
)
async def categories(
    service: Annotated[CategoryService, Depends(get_category_service)],
    parent_id: str | None = None,
    include_counts: bool = True,
) -> list[Category]:
    return await service.tree(parent_id, include_counts)
