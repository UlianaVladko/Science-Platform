"""Input DTOs and constrained application queries, independent of HTTP."""

from datetime import date
from typing import Annotated, Literal
from pydantic import Field, model_validator
from app.domain.models import (
    Model,
    PublicationType,
    ExternalIdentifier,
    ClusteringParameters,
    GraphType,
)

ID = Annotated[str, Field(min_length=1, max_length=200)]
Period = Literal["all", "2024", "2025", "2026"]


class Pagination(Model):
    page: int = Field(default=1, ge=1)
    page_size: int = Field(default=20, ge=1, le=100)


class ReferenceQuery(Pagination):
    direction: Literal["incoming", "outgoing"] = "outgoing"


class PublicationAuthorInput(Model):
    author_id: ID
    affiliation_ids: list[ID] | None = None


class PublicationCreate(Model):
    title: str = Field(min_length=1, max_length=1000)
    abstract: str | None = None
    publication_type: PublicationType
    publication_date: date | None = None
    language: str = Field(default="en", min_length=2, max_length=10)
    authors: list[PublicationAuthorInput] = Field(min_length=1)
    category_ids: list[ID] = Field(default_factory=list)
    external_identifiers: list[ExternalIdentifier] = Field(default_factory=list)
    journal: str | None = None
    publisher: str | None = None
    keywords: list[str] = Field(default_factory=list)
    model_config = {
        "extra": "forbid",
        "json_schema_extra": {
            "examples": [
                {
                    "title": "Scientific text analysis",
                    "publication_type": "article",
                    "authors": [{"author_id": "author-1"}],
                    "category_ids": ["cat-nlp"],
                    "language": "en",
                }
            ]
        },
    }


class PublicationUpdate(Model):
    title: str | None = Field(default=None, min_length=1, max_length=1000)
    abstract: str | None = None
    publication_type: PublicationType | None = None
    publication_date: date | None = None
    language: str | None = Field(default=None, min_length=2, max_length=10)
    authors: list[PublicationAuthorInput] | None = Field(default=None, min_length=1)
    category_ids: list[ID] | None = None
    external_identifiers: list[ExternalIdentifier] | None = None
    journal: str | None = None
    publisher: str | None = None
    keywords: list[str] | None = None

    @model_validator(mode="after")
    def validate_patch(self) -> "PublicationUpdate":
        if not self.model_fields_set:
            raise ValueError("PATCH must contain at least one field")
        nullable = {"abstract", "publication_date", "journal", "publisher"}
        for key in self.model_fields_set - nullable:
            if getattr(self, key) is None:
                raise ValueError(f"{key} cannot be null; omit it to leave unchanged")
        return self


class SearchFilters(Model):
    category: ID | None = None
    year_from: int | None = Field(default=None, ge=1000, le=2100)
    year_to: int | None = Field(default=None, ge=1000, le=2100)
    publication_type: list[PublicationType] | None = None
    language: str | None = None
    author: ID | None = None
    organization: ID | None = None
    journal: str | None = None
    terms: list[str] | None = None

    @model_validator(mode="after")
    def years(self) -> "SearchFilters":
        if self.year_from and self.year_to and self.year_from > self.year_to:
            raise ValueError("year_from must not exceed year_to")
        return self


class SearchRequest(Pagination):
    query: str | None = Field(default=None, max_length=2000)
    entity_type: Literal["publication", "author"] = "publication"
    mode: Literal["lexical", "semantic", "hybrid", "similar"] = "hybrid"
    similar_to_publication_id: ID | None = None
    filters: SearchFilters = Field(default_factory=SearchFilters)
    sort: Literal["relevance", "date", "citations", "views", "downloads"] = "relevance"
    include_facets: bool = False

    @model_validator(mode="after")
    def combinations(self) -> "SearchRequest":
        if self.mode == "similar":
            if self.entity_type != "publication" or not self.similar_to_publication_id:
                raise ValueError("similar requires publication entity and source ID")
        elif self.similar_to_publication_id:
            raise ValueError("source ID is only allowed in similar mode")
        if self.entity_type == "author" and self.sort not in {"relevance", "citations"}:
            raise ValueError("authors support relevance or citations sort")
        return self


Metric = Literal["publications", "added_documents", "citations", "term_frequency"]
Dimension = Literal[
    "publication_year",
    "created_date",
    "category",
    "language",
    "publication_type",
    "organization",
    "status",
    "source",
    "coauthors_count",
    "citation_year",
    "keyword",
    "concept",
    "lemma",
]
AGGREGATE_DIMENSIONS: dict[str, set[str]] = {
    "publications": {
        "publication_year",
        "category",
        "language",
        "publication_type",
        "organization",
        "status",
        "source",
        "coauthors_count",
    },
    "added_documents": {"created_date", "category"},
    "citations": {"citation_year", "publication_year", "category"},
    "term_frequency": {"keyword", "concept", "lemma", "publication_year", "category"},
}


class AggregateQuery(Model):
    metric: Metric = "publications"
    group_by: str = Field(
        default="publication_year",
        description="One or two unique comma-separated dimensions; see metric/group_by matrix in README.",
        pattern=r"^(publication_year|created_date|category|language|publication_type|organization|status|source|coauthors_count|citation_year|keyword|concept|lemma)(,(publication_year|created_date|category|language|publication_type|organization|status|source|coauthors_count|citation_year|keyword|concept|lemma))?$",
    )
    category: ID | None = None
    period: Period = "all"
    language: str | None = None
    publication_type: PublicationType | None = None
    author: ID | None = None
    organization: ID | None = None
    term: str | None = Field(default=None, min_length=1)

    @model_validator(mode="after")
    def combinations(self) -> "AggregateQuery":
        dims = self.dimensions()
        if not 1 <= len(dims) <= 2 or len(set(dims)) != len(dims):
            raise ValueError(
                "group_by requires one or two unique comma-separated dimensions"
            )
        if not set(dims) <= AGGREGATE_DIMENSIONS[self.metric]:
            raise ValueError("Unsupported metric/group_by combination")
        if (
            self.metric == "term_frequency"
            and "publication_year" in dims
            and not self.term
        ):
            raise ValueError("term is required for a term frequency timeline")
        if self.term is not None and self.metric != "term_frequency":
            raise ValueError("term requires term_frequency metric")
        return self

    def dimensions(self) -> list[str]:
        return self.group_by.split(",")


class RankingQuery(Model):
    entity: Literal["publication", "author", "keyword", "concept"] = "publication"
    metric: Literal["citations", "views", "downloads", "publications", "frequency"] = (
        "citations"
    )
    category: ID | None = None
    period: Period = "all"
    limit: int = Field(default=10, ge=1, le=100)

    @model_validator(mode="after")
    def combinations(self) -> "RankingQuery":
        allowed = {
            "publication": {"citations", "views", "downloads"},
            "author": {"citations", "publications"},
            "keyword": {"frequency"},
            "concept": {"frequency"},
        }
        if self.metric not in allowed[self.entity]:
            raise ValueError("Unsupported entity/metric combination")
        return self


class GraphQuery(Model):
    type: GraphType = "citation"
    root_id: ID | None = None
    max_depth: int = Field(default=2, ge=0, le=5)
    max_nodes: int = Field(default=50, ge=1, le=200)
    category: ID | None = None
    period: Period = "all"


class ProcessingParameters(Model):
    language: str | None = None
    force: bool = False


class TermParameters(Model):
    method: Literal["keybert", "tfidf"] = "keybert"
    top_k: int = Field(default=20, ge=1, le=100)
    language: str | None = None


class ProcessJobRequest(Model):
    operation: Literal["process_publications"]
    publication_ids: list[ID] = Field(min_length=1, max_length=1000)
    parameters: ProcessingParameters = Field(default_factory=ProcessingParameters)


class ExtractJobRequest(Model):
    operation: Literal["extract_terms"]
    publication_ids: list[ID] = Field(min_length=1, max_length=1000)
    parameters: TermParameters = Field(default_factory=TermParameters)


class ClusteringJobRequest(Model):
    operation: Literal["clustering"]
    publication_ids: list[ID] = Field(min_length=1, max_length=1000)
    parameters: ClusteringParameters = Field(default_factory=ClusteringParameters)


JobRequest = Annotated[
    ProcessJobRequest | ExtractJobRequest | ClusteringJobRequest,
    Field(discriminator="operation"),
]
