from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends, Query
from app.domain.models import SearchResponse, Suggestions
from app.schemas.requests import SearchRequest
from app.application.services import SearchService
from app.api.dependencies import get_search_service

router = APIRouter(tags=["Search"])


@router.post(
    "/search",
    response_model=SearchResponse,
    summary="Search publications or authors",
    description="Select deterministic fixtures in lexical, semantic, hybrid or publication-only similar mode. No actual search algorithms run; all filters apply before pagination.",
    openapi_extra={
        "requestBody": {
            "content": {
                "application/json": {
                    "example": {
                        "query": "text",
                        "entity_type": "publication",
                        "mode": "hybrid",
                        "filters": {},
                        "page": 1,
                        "page_size": 20,
                        "include_facets": True,
                    }
                }
            }
        }
    },
    responses=success_response("search", 200),
)
async def search(
    command: SearchRequest,
    service: Annotated[SearchService, Depends(get_search_service)],
) -> SearchResponse:
    return await service.search(command)


@router.get(
    "/search/suggest",
    response_model=Suggestions,
    summary="Get search suggestions",
    description="Return matching publication titles, author names and scientific terms from fixtures.",
    responses=success_response("suggest", 200),
)
async def suggest(
    service: Annotated[SearchService, Depends(get_search_service)],
    q: Annotated[str, Query(min_length=1, max_length=200)],
    limit: Annotated[int, Query(ge=1, le=50)] = 10,
) -> Suggestions:
    return await service.suggest(q, limit)
