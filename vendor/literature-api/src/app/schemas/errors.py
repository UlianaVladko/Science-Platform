from app.domain.models import Model


class ValidationIssue(Model):
    location: list[str | int]
    message: str
    type: str


class ErrorDetail(Model):
    code: str
    message: str
    details: list[ValidationIssue] | None = None
    request_id: str


class ErrorResponse(Model):
    error: ErrorDetail


ERROR_RESPONSES = {
    status: {
        "model": ErrorResponse,
        "description": description,
        "content": {
            "application/json": {
                "example": {
                    "error": {
                        "code": code,
                        "message": description,
                        "details": None,
                        "request_id": "req-example",
                    }
                }
            }
        },
    }
    for status, code, description in [
        (400, "BAD_REQUEST", "Invalid request"),
        (404, "PUBLICATION_NOT_FOUND", "Object not found"),
        (409, "IDENTIFIER_CONFLICT", "Operation conflict"),
        (422, "VALIDATION_ERROR", "Input validation failed"),
        (500, "INTERNAL_ERROR", "Internal server error"),
    ]
}
