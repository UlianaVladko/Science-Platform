import logging
from uuid import uuid4
from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException
from starlette.types import ASGIApp, Scope, Receive, Send, Message
from app.domain.exceptions import ApplicationError
from app.schemas.errors import ErrorResponse, ErrorDetail, ValidationIssue

logger = logging.getLogger(__name__)


class RequestIDMiddleware:
    """ASGI middleware also adds the identifier to error responses."""

    def __init__(self, app: ASGIApp) -> None:
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return
        scope.setdefault("state", {})["request_id"] = "req-" + uuid4().hex

        async def send_with_id(message: Message) -> None:
            if message["type"] == "http.response.start":
                message["headers"] = [
                    (k, v)
                    for k, v in message.get("headers", [])
                    if k.lower() != b"x-request-id"
                ] + [(b"x-request-id", scope["state"]["request_id"].encode())]
            await send(message)

        await self.app(scope, receive, send_with_id)


def response(
    request: Request,
    status: int,
    code: str,
    message: str,
    details: list[ValidationIssue] | None = None,
) -> JSONResponse:
    request_id = getattr(request.state, "request_id", "req-" + uuid4().hex)
    body = ErrorResponse(
        error=ErrorDetail(
            code=code, message=message, details=details, request_id=request_id
        )
    )
    return JSONResponse(
        status_code=status,
        content=body.model_dump(mode="json"),
        headers={"X-Request-ID": request_id},
    )


def register_handlers(app: FastAPI) -> None:
    @app.exception_handler(ApplicationError)
    async def application_error(
        request: Request, exc: ApplicationError
    ) -> JSONResponse:
        status = {"not_found": 404, "conflict": 409, "bad_request": 400}.get(
            exc.kind, 500
        )
        return response(request, status, exc.code, exc.message)

    @app.exception_handler(RequestValidationError)
    async def validation_error(
        request: Request, exc: RequestValidationError
    ) -> JSONResponse:
        issues = [
            ValidationIssue(location=list(e["loc"]), message=e["msg"], type=e["type"])
            for e in exc.errors()
        ]
        return response(
            request, 422, "VALIDATION_ERROR", "Input validation failed", issues
        )

    @app.exception_handler(HTTPException)
    async def http_error(request: Request, exc: HTTPException) -> JSONResponse:
        return response(
            request, exc.status_code, f"HTTP_{exc.status_code}", str(exc.detail)
        )

    @app.exception_handler(Exception)
    async def unexpected_error(request: Request, exc: Exception) -> JSONResponse:
        logger.error(
            "Unhandled API error %s",
            getattr(request.state, "request_id", None),
            exc_info=exc,
        )
        return response(request, 500, "INTERNAL_ERROR", "Internal server error")
