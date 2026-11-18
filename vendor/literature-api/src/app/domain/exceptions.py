class ApplicationError(Exception):
    """Application failure; the HTTP layer maps kind to a status code."""

    def __init__(self, code: str, message: str, kind: str = "not_found") -> None:
        self.code = code
        self.message = message
        self.kind = kind
        super().__init__(message)


def not_found(entity: str) -> ApplicationError:
    return ApplicationError(
        f"{entity.upper()}_NOT_FOUND", f"{entity.title()} not found"
    )
