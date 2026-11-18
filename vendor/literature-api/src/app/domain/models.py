"""Transport-independent read models; no FastAPI or adapter dependencies."""

from datetime import date, datetime
from enum import StrEnum
from typing import Annotated, Generic, Literal, TypeVar
from pydantic import BaseModel, ConfigDict, Field


class Model(BaseModel):
    model_config = ConfigDict(extra="forbid", validate_assignment=True)


class PublicationType(StrEnum):
    article = "article"
    journal_publication = "journal_publication"
    book = "book"
    monograph = "monograph"
    preprint = "preprint"
    other = "other"


class PublicationProcessingStatus(StrEnum):
    pending = "pending"
    processing = "processing"
    completed = "completed"
    failed = "failed"


class JobStatus(StrEnum):
    pending = "pending"
    running = "running"
    completed = "completed"
    failed = "failed"


FileKind = Literal["raw", "markdown"]
Operation = Literal["process_publications", "extract_terms", "clustering"]
GraphType = Literal["citation", "coauthor", "concept"]
T = TypeVar("T")


class Page(Model, Generic[T]):
    items: list[T]
    total: int = Field(ge=0)
    page: int = Field(ge=1)
    page_size: int = Field(ge=1, le=100)


class Organization(Model):
    id: str
    name: str


class Category(Model):
    id: str
    slug: str
    name: str
    description: str | None = None
    parent_id: str | None = None
    publications_count: int | None = Field(default=None, ge=0)
    children: list["Category"] = Field(default_factory=list)


class AuthorSummary(Model):
    id: str
    name: str
    orcid: str | None = None
    affiliations: list[Organization] = Field(default_factory=list)


class Author(AuthorSummary):
    categories: list[Category] = Field(default_factory=list)
    publications_count: int = Field(default=0, ge=0)
    citations_count: int = Field(default=0, ge=0)
    h_index: int | None = Field(default=None, ge=0)


class PublicationMetrics(Model):
    citations: int = Field(default=0, ge=0)
    views: int = Field(default=0, ge=0)
    downloads: int = Field(default=0, ge=0)


class ExternalIdentifier(Model):
    scheme: Literal["doi", "isbn", "arxiv", "other"]
    value: str = Field(min_length=1)


class ScientificConcept(Model):
    id: str
    name: str


class PublicationFile(Model):
    kind: FileKind
    content_type: str
    size_bytes: int = Field(ge=0)


class PublicationSummary(Model):
    id: str
    title: str
    abstract: str | None = None
    publication_type: PublicationType
    publication_date: date | None = None
    language: str
    authors: list[AuthorSummary]
    metrics: PublicationMetrics = Field(default_factory=PublicationMetrics)


class Publication(PublicationSummary):
    external_identifiers: list[ExternalIdentifier] = Field(default_factory=list)
    journal: str | None = None
    publisher: str | None = None
    categories: list[Category] = Field(default_factory=list)
    keywords: list[str] = Field(default_factory=list)
    concepts: list[ScientificConcept] = Field(default_factory=list)
    files: list[PublicationFile] = Field(default_factory=list)
    processing_status: PublicationProcessingStatus = PublicationProcessingStatus.pending


class Reference(Model):
    citing_publication_id: str
    cited_publication_id: str | None
    raw_reference: str
    publication: PublicationSummary | None = None


class ViewCount(Model):
    publication_id: str
    views: int = Field(ge=0)


class FileAccepted(Model):
    publication_id: str
    file: PublicationFile
    accepted: bool = True


class FileContent(Model):
    content: bytes
    content_type: str
    filename: str


class PublicationHit(Model):
    entity_type: Literal["publication"] = "publication"
    score: float = Field(ge=0)
    entity: PublicationSummary
    snippets: list[str] = Field(default_factory=list)


class AuthorHit(Model):
    entity_type: Literal["author"] = "author"
    score: float = Field(ge=0)
    entity: AuthorSummary
    snippets: list[str] = Field(default_factory=list)


SearchResult = Annotated[PublicationHit | AuthorHit, Field(discriminator="entity_type")]


class FacetBucket(Model):
    value: str
    count: int = Field(ge=0)


class SearchFacet(Model):
    field: str
    buckets: list[FacetBucket]


class SearchResponse(Page[SearchResult]):
    facets: list[SearchFacet] | None = None


class Suggestion(Model):
    type: Literal["publication", "author", "term"]
    text: str
    matches_count: int | None = Field(default=None, ge=0)


class Suggestions(Model):
    items: list[Suggestion]


class CorpusSummary(Model):
    total_publications: int = Field(ge=0)
    total_authors: int = Field(ge=0)
    total_citations: int = Field(ge=0)
    total_downloads: int = Field(ge=0)
    avg_citations_per_publication: float = Field(ge=0)
    publications_this_period: int = Field(ge=0)
    new_authors_this_period: int = Field(ge=0)


class AggregatePoint(Model):
    dimensions: dict[str, str]
    value: float


class AggregateResult(Model):
    metric: str
    group_by: list[str]
    points: list[AggregatePoint]
    data_as_of: datetime


class RankingObject(Model):
    entity_type: Literal["publication", "author", "keyword", "concept"]
    id: str
    label: str


class RankingItem(Model):
    position: int = Field(ge=1)
    object: RankingObject
    value: float = Field(ge=0)


class Ranking(Model):
    metric: str
    items: list[RankingItem]


class GraphNode(Model):
    id: str
    type: Literal["publication", "author", "concept"]
    label: str


class GraphEdge(Model):
    source: str
    target: str
    type: Literal["cites", "coauthor", "cooccurs"]
    weight: float | None = Field(default=None, ge=0)


class Graph(Model):
    nodes: list[GraphNode]
    edges: list[GraphEdge]
    truncated: bool
    data_as_of: datetime


class JobResult(Model):
    processed_publications: int = Field(ge=0)
    run_id: str | None = None


class JobError(Model):
    code: str
    message: str


class AnalysisJob(Model):
    job_id: str
    operation: Operation
    status: JobStatus
    progress: float = Field(ge=0, le=100)
    created_at: datetime
    completed_at: datetime | None = None
    result: JobResult | None = None
    error: JobError | None = None
    run_id: str | None = None


class ClusteringParameters(Model):
    algorithm: Literal["kmeans", "hdbscan"] = "kmeans"
    feature_source: Literal["embeddings", "keywords"] = "embeddings"
    n_clusters: int = Field(default=2, ge=2, le=100)


class ClusterMetrics(Model):
    silhouette: float | None = Field(default=None, ge=-1, le=1)
    clusters_count: int = Field(ge=0)


class Cluster(Model):
    cluster_id: str
    size: int = Field(ge=0)
    publication_ids: list[str]
    keywords: list[str] | None = None


class ClusteringRun(Model):
    run_id: str
    algorithm: Literal["kmeans", "hdbscan"]
    parameters: ClusteringParameters
    status: JobStatus
    processed_publications: int = Field(ge=0)
    metrics: ClusterMetrics


class ClusteringRunDetail(ClusteringRun):
    clusters: list[Cluster]
