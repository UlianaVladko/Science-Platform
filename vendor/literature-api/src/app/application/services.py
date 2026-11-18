"""Thin application facade. Real implementations can replace ports without HTTP changes."""

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
from app.application.ports import (
    PublicationPort,
    SearchPort,
    AuthorPort,
    CategoryPort,
    AnalyticsPort,
    AnalysisJobPort,
)


class PublicationService:
    def __init__(self, port: PublicationPort) -> None:
        self.port = port

    async def get(self, publication_id: str) -> Publication:
        return await self.port.get(publication_id)

    async def create(self, command: PublicationCreate) -> Publication:
        return await self.port.create(command)

    async def update(
        self, publication_id: str, command: PublicationUpdate
    ) -> Publication:
        return await self.port.update(publication_id, command)

    async def delete(self, publication_id: str) -> None:
        return await self.port.delete(publication_id)

    async def put_file(
        self, publication_id: str, kind: FileKind, content: bytes, content_type: str
    ) -> FileAccepted:
        return await self.port.put_file(publication_id, kind, content, content_type)

    async def get_file(self, publication_id: str, kind: FileKind) -> FileContent:
        return await self.port.get_file(publication_id, kind)

    async def references(
        self, publication_id: str, direction: str, pagination: Pagination
    ) -> Page[Reference]:
        return await self.port.references(publication_id, direction, pagination)

    async def view(self, publication_id: str) -> ViewCount:
        return await self.port.view(publication_id)


class SearchService:
    def __init__(self, port: SearchPort) -> None:
        self.port = port

    async def search(self, query: SearchRequest) -> SearchResponse:
        return await self.port.search(query)

    async def suggest(self, query: str, limit: int) -> Suggestions:
        return await self.port.suggest(query, limit)


class AuthorService:
    def __init__(self, port: AuthorPort) -> None:
        self.port = port

    async def get(self, author_id: str) -> Author:
        return await self.port.get(author_id)


class CategoryService:
    def __init__(self, port: CategoryPort) -> None:
        self.port = port

    async def tree(self, parent_id: str | None, include_counts: bool) -> list[Category]:
        return await self.port.tree(parent_id, include_counts)


class AnalyticsService:
    def __init__(self, port: AnalyticsPort) -> None:
        self.port = port

    async def summary(self) -> CorpusSummary:
        return await self.port.summary()

    async def aggregates(self, query: AggregateQuery) -> AggregateResult:
        return await self.port.aggregates(query)

    async def top(self, query: RankingQuery) -> Ranking:
        return await self.port.top(query)

    async def graph(self, query: GraphQuery) -> Graph:
        return await self.port.graph(query)


class AnalysisJobService:
    def __init__(self, port: AnalysisJobPort) -> None:
        self.port = port

    async def start(self, command: JobRequest) -> AnalysisJob:
        return await self.port.start(command)

    async def get(self, job_id: str) -> AnalysisJob:
        return await self.port.get(job_id)

    async def runs(self, pagination: Pagination) -> Page[ClusteringRun]:
        return await self.port.runs(pagination)

    async def run(self, run_id: str) -> ClusteringRunDetail:
        return await self.port.run(run_id)
