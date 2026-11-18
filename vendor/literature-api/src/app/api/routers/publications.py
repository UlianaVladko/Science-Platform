from app.schemas.examples import success_response
from typing import Annotated
from fastapi import APIRouter, Depends, Query, Path, Body, Response, Header
from app.domain.models import (
    FileAccepted,
    FileKind,
    Page,
    Publication,
    Reference,
    ViewCount,
)
from app.schemas.requests import PublicationCreate, PublicationUpdate, ReferenceQuery
from app.application.services import PublicationService
from app.api.dependencies import get_publication_service

router = APIRouter(tags=["Publications"])
PublicationID = Annotated[
    str, Path(min_length=1, description="Opaque publication ID", examples=["pub-1"])
]
Service = Annotated[PublicationService, Depends(get_publication_service)]


@router.get(
    "/publications/{id}",
    response_model=Publication,
    summary="Get publication",
    description="Full metadata, ordered authors, categories, concepts, metrics, files and processing status.",
    responses=success_response("get_publication", 200),
)
async def get_publication(id: PublicationID, service: Service) -> Publication:
    return await service.get(id)


@router.post(
    "/publications",
    status_code=201,
    response_model=Publication,
    summary="Create publication",
    description="Create an in-memory publication. Authors and categories must exist; external identifiers must be unique.",
    responses=success_response("create_publication", 201),
)
async def create_publication(
    command: PublicationCreate, service: Service
) -> Publication:
    return await service.create(command)


@router.patch(
    "/publications/{id}",
    response_model=Publication,
    summary="Update publication metadata",
    description="Omitted fields remain unchanged. Null clears nullable metadata; server-controlled fields are rejected.",
    responses=success_response("update_publication", 200),
)
async def update_publication(
    id: PublicationID, command: PublicationUpdate, service: Service
) -> Publication:
    return await service.update(id, command)


@router.delete(
    "/publications/{id}",
    status_code=204,
    response_class=Response,
    summary="Delete publication",
    description="Remove the in-memory publication and its files; incoming references become unresolved.",
)
async def delete_publication(id: PublicationID, service: Service) -> Response:
    await service.delete(id)
    return Response(status_code=204)


@router.put(
    "/publications/{id}/files/{kind}",
    response_model=FileAccepted,
    summary="Accept or replace publication file",
    description="Binary body: application/pdf for raw, text/markdown (UTF-8) for markdown. Maximum 5 MiB. Stored only in memory.",
    openapi_extra={
        "requestBody": {
            "required": True,
            "content": {
                "application/pdf": {"schema": {"type": "string", "format": "binary"}},
                "text/markdown": {"schema": {"type": "string", "format": "binary"}},
            },
        }
    },
    responses=success_response("put_file", 200),
)
async def put_file(
    id: PublicationID,
    kind: FileKind,
    service: Service,
    content_type: Annotated[str | None, Header()] = None,
    content: bytes = Body(media_type="application/pdf"),
) -> FileAccepted:
    return await service.put_file(id, kind, content, content_type or "")


@router.get(
    "/publications/{id}/files/{kind}",
    response_class=Response,
    summary="Download publication file",
    description="Return prepared PDF or Markdown bytes, or the last uploaded in-memory replacement.",
    responses={
        200: {
            "description": "Publication file",
            "content": {
                "application/pdf": {"schema": {"type": "string", "format": "binary"}},
                "text/markdown": {"schema": {"type": "string", "format": "binary"}},
            },
        }
    },
)
async def get_file(id: PublicationID, kind: FileKind, service: Service) -> Response:
    result = await service.get_file(id, kind)
    return Response(
        content=result.content,
        media_type=result.content_type,
        headers={"Content-Disposition": f'attachment; filename="{result.filename}"'},
    )


@router.get(
    "/publications/{id}/references",
    response_model=Page[Reference],
    summary="Get bibliography or citations",
    description="Outgoing references target cited works; incoming references identify citing works. Unresolved references have a null cited_publication_id.",
    responses=success_response("references", 200),
)
async def references(
    id: PublicationID, service: Service, query: Annotated[ReferenceQuery, Query()]
) -> Page[Reference]:
    return await service.references(id, query.direction, query)


@router.post(
    "/publications/{id}/view",
    response_model=ViewCount,
    summary="Register publication view",
    description="Increment the fixture view count once for every call. No deduplication is implemented.",
    responses=success_response("view", 200),
)
async def view(id: PublicationID, service: Service) -> ViewCount:
    return await service.view(id)
