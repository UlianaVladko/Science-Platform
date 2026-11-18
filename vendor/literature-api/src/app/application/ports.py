"""Application-owned capabilities; no HTTP or implementation dependencies."""

from typing import Protocol
from app.domain.models import (
    Publication,
    FileKind,
    FileAccepted,
    FileContent,
    Page,
    Reference,
    ViewCount,
    SearchResponse,
    Suggestions,
    Author,
    Category,
    CorpusSummary,
    AggregateResult,
    Ranking,
    Graph,
    AnalysisJob,
    ClusteringRun,
    ClusteringRunDetail,
)
from app.schemas.requests import (
    PublicationCreate,
    PublicationUpdate,
    Pagination,
    SearchRequest,
    AggregateQuery,
    RankingQuery,
    GraphQuery,
    JobRequest,
)


class PublicationPort(Protocol):
    async def get(self, publication_id: str) -> Publication: ...
    async def create(self, command: PublicationCreate) -> Publication: ...
    async def update(
        self, publication_id: str, command: PublicationUpdate
    ) -> Publication: ...
    async def delete(self, publication_id: str) -> None: ...
    async def put_file(
        self, publication_id: str, kind: FileKind, content: bytes, content_type: str
    ) -> FileAccepted: ...
    async def get_file(self, publication_id: str, kind: FileKind) -> FileContent: ...
    async def references(
        self, publication_id: str, direction: str, pagination: Pagination
    ) -> Page[Reference]: ...
    async def view(self, publication_id: str) -> ViewCount: ...


class SearchPort(Protocol):
    async def search(self, query: SearchRequest) -> SearchResponse: ...
    async def suggest(self, query: str, limit: int) -> Suggestions: ...


class AuthorPort(Protocol):
    async def get(self, author_id: str) -> Author: ...


class CategoryPort(Protocol):
    async def tree(
        self, parent_id: str | None, include_counts: bool
    ) -> list[Category]: ...


class AnalyticsPort(Protocol):
    async def summary(self) -> CorpusSummary: ...
    async def aggregates(self, query: AggregateQuery) -> AggregateResult: ...
    async def top(self, query: RankingQuery) -> Ranking: ...
    async def graph(self, query: GraphQuery) -> Graph: ...


class AnalysisJobPort(Protocol):
    async def start(self, command: JobRequest) -> AnalysisJob: ...
    async def get(self, job_id: str) -> AnalysisJob: ...
    async def runs(self, pagination: Pagination) -> Page[ClusteringRun]: ...
    async def run(self, run_id: str) -> ClusteringRunDetail: ...
