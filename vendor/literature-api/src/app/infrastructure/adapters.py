"""Deterministic mock adapters. Filtering here is fixture selection, not a search engine."""

from collections import Counter, deque
from datetime import datetime, timezone
from typing import TypeVar
from app.application.ports import (
    PublicationPort,
    SearchPort,
    AuthorPort,
    CategoryPort,
    AnalyticsPort,
    AnalysisJobPort,
)
from app.domain.models import (
    AggregatePoint,
    AggregateResult,
    AnalysisJob,
    Author,
    AuthorHit,
    AuthorSummary,
    Category,
    ClusterMetrics,
    ClusteringRun,
    ClusteringRunDetail,
    CorpusSummary,
    ExternalIdentifier,
    FacetBucket,
    FileAccepted,
    FileContent,
    FileKind,
    Graph,
    GraphEdge,
    GraphNode,
    Page,
    Publication,
    PublicationFile,
    PublicationHit,
    PublicationSummary,
    Ranking,
    RankingItem,
    RankingObject,
    Reference,
    SearchFacet,
    SearchResponse,
    Suggestion,
    Suggestions,
    ViewCount,
)
from app.domain.exceptions import ApplicationError, not_found
from app.schemas.requests import (
    PublicationCreate,
    PublicationUpdate,
    Pagination,
    SearchRequest,
    SearchFilters,
    AggregateQuery,
    RankingQuery,
    GraphQuery,
    JobRequest,
)
from app.infrastructure.store import MockStore, AS_OF

T = TypeVar("T")


def paginate(items: list[T], pagination: Pagination) -> Page[T]:
    start = (pagination.page - 1) * pagination.page_size
    return Page(
        items=items[start : start + pagination.page_size],
        total=len(items),
        page=pagination.page,
        page_size=pagination.page_size,
    )


def summary(p: Publication) -> PublicationSummary:
    return PublicationSummary.model_validate(
        p.model_dump(include=set(PublicationSummary.model_fields))
    )


def author_summary(a: Author) -> AuthorSummary:
    return AuthorSummary.model_validate(
        a.model_dump(include=set(AuthorSummary.model_fields))
    )


class BaseAdapter:
    def __init__(self, store: MockStore) -> None:
        self.store = store

    def publication(self, publication_id: str) -> Publication:
        if publication_id not in self.store.publications:
            raise not_found("publication")
        return self.store.publications[publication_id]

    def category_ids(self, category: str) -> set[str]:
        if category not in self.store.categories:
            raise not_found("category")
        result = {category}
        for c in self.store.categories.values():
            if c.parent_id == category:
                result |= self.category_ids(c.id)
        return result

    def select(self, filters: SearchFilters, period: str = "all") -> list[Publication]:
        category_ids = self.category_ids(filters.category) if filters.category else None
        result = []
        for p in self.store.publications.values():
            year = p.publication_date.year if p.publication_date else None
            if period != "all" and str(year) != period:
                continue
            if category_ids and not any(c.id in category_ids for c in p.categories):
                continue
            if filters.year_from and (year is None or year < filters.year_from):
                continue
            if filters.year_to and (year is None or year > filters.year_to):
                continue
            if (
                filters.publication_type
                and p.publication_type not in filters.publication_type
            ):
                continue
            if filters.language and p.language != filters.language:
                continue
            if filters.author and not any(a.id == filters.author for a in p.authors):
                continue
            if filters.organization and not any(
                o.id == filters.organization for a in p.authors for o in a.affiliations
            ):
                continue
            if filters.journal and p.journal != filters.journal:
                continue
            if filters.terms and not set(t.casefold() for t in filters.terms) <= {
                t.casefold() for t in p.keywords + [c.name for c in p.concepts]
            }:
                continue
            result.append(p)
        return result


class MockPublicationAdapter(BaseAdapter, PublicationPort):
    async def get(self, publication_id: str) -> Publication:
        return self.publication(publication_id).model_copy(deep=True)

    def metadata(
        self, command: PublicationCreate | PublicationUpdate
    ) -> dict[str, object]:
        data = command.model_dump(exclude_unset=isinstance(command, PublicationUpdate))
        if "authors" in data:
            authors = command.authors or []
            ids = [a.author_id for a in authors]
            if any(i not in self.store.authors for i in ids):
                raise not_found("author")
            if len(set(ids)) != len(ids):
                raise ApplicationError(
                    "DUPLICATE_AUTHOR", "Duplicate author IDs", "bad_request"
                )
            organizations = {
                o.id: o for a in self.store.authors.values() for o in a.affiliations
            }
            result = []
            for entry in authors:
                value = author_summary(self.store.authors[entry.author_id]).model_copy(
                    deep=True
                )
                if entry.affiliation_ids is not None:
                    if any(id not in organizations for id in entry.affiliation_ids):
                        raise not_found("organization")
                    value.affiliations = [
                        organizations[id] for id in entry.affiliation_ids
                    ]
                result.append(value)
            data["authors"] = result
        if "category_ids" in data:
            ids = data.pop("category_ids")
            if any(i not in self.store.categories for i in ids):
                raise not_found("category")
            data["categories"] = [self.store.categories[i] for i in ids]
        return data

    def check_identifiers(
        self, identifiers: list[ExternalIdentifier], exclude_id: str | None = None
    ) -> None:
        pairs = {(x.scheme, x.value.casefold()) for x in identifiers}
        if len(pairs) != len(identifiers):
            raise ApplicationError(
                "IDENTIFIER_CONFLICT", "Duplicate external identifier", "conflict"
            )
        for p in self.store.publications.values():
            if p.id != exclude_id and pairs & {
                (x.scheme, x.value.casefold()) for x in p.external_identifiers
            }:
                raise ApplicationError(
                    "IDENTIFIER_CONFLICT",
                    "External identifier already exists",
                    "conflict",
                )

    async def create(self, command: PublicationCreate) -> Publication:
        self.check_identifiers(command.external_identifiers)
        data = self.metadata(command)
        self.store.publication_sequence += 1
        p = Publication(id=f"pub-{self.store.publication_sequence}", **data)
        self.store.publications[p.id] = p
        return p.model_copy(deep=True)

    async def update(
        self, publication_id: str, command: PublicationUpdate
    ) -> Publication:
        current = self.publication(publication_id)
        if command.external_identifiers is not None:
            self.check_identifiers(command.external_identifiers, publication_id)
        data = current.model_dump()
        data.update(self.metadata(command))
        p = Publication.model_validate(data)
        self.store.publications[p.id] = p
        return p.model_copy(deep=True)

    async def delete(self, publication_id: str) -> None:
        self.publication(publication_id)
        del self.store.publications[publication_id]
        self.store.files = {
            k: v for k, v in self.store.files.items() if k[0] != publication_id
        }
        self.store.references = [
            r
            for r in self.store.references
            if r.citing_publication_id != publication_id
        ]
        for r in self.store.references:
            if r.cited_publication_id == publication_id:
                r.cited_publication_id = None
                r.publication = None

    async def put_file(
        self, publication_id: str, kind: FileKind, content: bytes, content_type: str
    ) -> FileAccepted:
        p = self.publication(publication_id)
        expected = "application/pdf" if kind == "raw" else "text/markdown"
        if content_type.split(";")[0].strip().lower() != expected:
            raise ApplicationError(
                "INVALID_CONTENT_TYPE", f"Expected {expected}", "bad_request"
            )
        if not content:
            raise ApplicationError(
                "EMPTY_FILE", "File body must not be empty", "bad_request"
            )
        if len(content) > 5 * 1024 * 1024:
            raise ApplicationError(
                "FILE_TOO_LARGE", "Maximum file size is 5 MiB", "bad_request"
            )
        if kind == "raw" and not content.startswith(b"%PDF-"):
            raise ApplicationError(
                "INVALID_FILE", "Expected PDF signature", "bad_request"
            )
        if kind == "markdown":
            try:
                content.decode("utf-8")
            except UnicodeDecodeError:
                raise ApplicationError(
                    "INVALID_FILE", "Markdown must be UTF-8", "bad_request"
                ) from None
        info = PublicationFile(
            kind=kind, content_type=expected, size_bytes=len(content)
        )
        self.store.files[publication_id, kind] = content
        p.files = [f for f in p.files if f.kind != kind] + [info]
        return FileAccepted(publication_id=publication_id, file=info)

    async def get_file(self, publication_id: str, kind: FileKind) -> FileContent:
        self.publication(publication_id)
        if (publication_id, kind) not in self.store.files:
            raise not_found("file")
        return FileContent(
            content=self.store.files[publication_id, kind],
            content_type="application/pdf" if kind == "raw" else "text/markdown",
            filename="publication.pdf" if kind == "raw" else "publication.md",
        )

    async def references(
        self, publication_id: str, direction: str, pagination: Pagination
    ) -> Page[Reference]:
        self.publication(publication_id)
        result = []
        for ref in self.store.references:
            if (
                direction == "outgoing" and ref.citing_publication_id == publication_id
            ) or (
                direction == "incoming" and ref.cited_publication_id == publication_id
            ):
                item = ref.model_copy(deep=True)
                target = (
                    item.cited_publication_id
                    if direction == "outgoing"
                    else item.citing_publication_id
                )
                item.publication = (
                    summary(self.store.publications[target])
                    if target in self.store.publications
                    else None
                )
                result.append(item)
        return paginate(result, pagination)

    async def view(self, publication_id: str) -> ViewCount:
        p = self.publication(publication_id)
        p.metrics.views += 1
        return ViewCount(publication_id=publication_id, views=p.metrics.views)


class MockCategoryAdapter(BaseAdapter, CategoryPort):
    async def tree(self, parent_id: str | None, include_counts: bool) -> list[Category]:
        if parent_id is not None:
            self.category_ids(parent_id)

        def build(c: Category) -> Category:
            value = c.model_copy(deep=True)
            ids = self.category_ids(c.id)
            value.publications_count = (
                sum(
                    any(x.id in ids for x in p.categories)
                    for p in self.store.publications.values()
                )
                if include_counts
                else None
            )
            value.children = [
                build(child)
                for child in self.store.categories.values()
                if child.parent_id == c.id
            ]
            return value

        return [
            build(c) for c in self.store.categories.values() if c.parent_id == parent_id
        ]


class MockAuthorAdapter(BaseAdapter, AuthorPort):
    async def get(self, author_id: str) -> Author:
        if author_id not in self.store.authors:
            raise not_found("author")
        a = self.store.authors[author_id].model_copy(deep=True)
        pubs = self.select(SearchFilters(author=author_id))
        a.publications_count = len(pubs)
        a.citations_count = sum(p.metrics.citations for p in pubs)
        a.categories = list({c.id: c for p in pubs for c in p.categories}.values())
        counts = sorted((p.metrics.citations for p in pubs), reverse=True)
        a.h_index = max((i for i, c in enumerate(counts, 1) if c >= i), default=0)
        return a


class MockSearchAdapter(BaseAdapter, SearchPort):
    async def search(self, query: SearchRequest) -> SearchResponse:
        pubs = self.select(query.filters)
        if query.mode == "similar":
            source = self.publication(query.similar_to_publication_id)
            pubs = [p for p in pubs if p.id != source.id]
        text = (query.query or "").casefold()
        if query.entity_type == "publication":
            pubs = [
                p
                for p in pubs
                if not text
                or text
                in (
                    p.title
                    + " "
                    + (p.abstract or "")
                    + " "
                    + " ".join(a.name for a in p.authors)
                ).casefold()
            ]
            if query.sort == "date":
                pubs.sort(key=lambda p: str(p.publication_date or ""), reverse=True)
            elif query.sort != "relevance":
                pubs.sort(key=lambda p: getattr(p.metrics, query.sort), reverse=True)
            hits = [
                PublicationHit(
                    score=1 / (i + 1),
                    entity=summary(p),
                    snippets=[p.abstract] if text and p.abstract else [],
                )
                for i, p in enumerate(pubs)
            ]
        else:
            eligible = {a.id for p in pubs for a in p.authors}
            has_filters = any(
                v is not None for v in query.filters.model_dump().values()
            )
            authors = [
                a
                for a in self.store.authors.values()
                if (not has_filters or a.id in eligible)
                and (not text or text in a.name.casefold())
            ]
            if query.sort == "citations":
                authors.sort(
                    key=lambda a: sum(
                        p.metrics.citations
                        for p in self.store.publications.values()
                        if any(x.id == a.id for x in p.authors)
                    ),
                    reverse=True,
                )
            hits = [
                AuthorHit(score=1 / (i + 1), entity=author_summary(a))
                for i, a in enumerate(authors)
            ]
        page = paginate(hits, query)
        facets = None
        if query.include_facets:
            counts = (
                Counter(p.language for p in pubs)
                if query.entity_type == "publication"
                else Counter(o.name for a in authors for o in a.affiliations)
            )
            facets = [
                SearchFacet(
                    field="language"
                    if query.entity_type == "publication"
                    else "organization",
                    buckets=[
                        FacetBucket(value=k, count=v) for k, v in sorted(counts.items())
                    ],
                )
            ]
        return SearchResponse(**page.model_dump(), facets=facets)

    async def suggest(self, query: str, limit: int) -> Suggestions:
        entries = (
            [("publication", p.title) for p in self.store.publications.values()]
            + [("author", a.name) for a in self.store.authors.values()]
            + [
                ("term", t)
                for t in sorted(
                    {
                        c.name
                        for p in self.store.publications.values()
                        for c in p.concepts
                    }
                )
            ]
        )
        return Suggestions(
            items=[
                Suggestion(type=t, text=s, matches_count=1)
                for t, s in entries
                if query.casefold() in s.casefold()
            ][:limit]
        )


class MockAnalyticsAdapter(BaseAdapter, AnalyticsPort):
    def selected(
        self, query: AggregateQuery | RankingQuery | GraphQuery
    ) -> list[Publication]:
        return self.select(
            SearchFilters(
                category=query.category,
                language=getattr(query, "language", None),
                author=getattr(query, "author", None),
                organization=getattr(query, "organization", None),
                publication_type=[query.publication_type]
                if getattr(query, "publication_type", None)
                else None,
            ),
            query.period,
        )

    async def summary(self) -> CorpusSummary:
        pubs = list(self.store.publications.values())
        citations = sum(p.metrics.citations for p in pubs)
        return CorpusSummary(
            total_publications=len(pubs),
            total_authors=len(self.store.authors),
            total_citations=citations,
            total_downloads=sum(p.metrics.downloads for p in pubs),
            avg_citations_per_publication=citations / len(pubs) if pubs else 0,
            publications_this_period=sum(
                p.publication_date is not None and p.publication_date.year == 2026
                for p in pubs
            ),
            new_authors_this_period=1,
        )

    async def aggregates(self, query: AggregateQuery) -> AggregateResult:
        # Fixed observation values for metadata not present on publications (no NLP/date inference).
        from itertools import product

        counts: Counter[tuple[str, ...]] = Counter()
        dims = query.dimensions()
        for p in self.selected(query):
            values = {
                "publication_year": [
                    str(p.publication_date.year) if p.publication_date else "unknown"
                ],
                "created_date": ["2026-11-01"],
                "category": [c.name for c in p.categories],
                "language": [p.language],
                "publication_type": [p.publication_type.value],
                "organization": sorted(
                    {o.name for a in p.authors for o in a.affiliations}
                ),
                "status": [p.processing_status.value],
                "source": ["demo"],
                "coauthors_count": [str(len(p.authors))],
                "citation_year": ["2026"],
                "keyword": p.keywords,
                "concept": [c.name for c in p.concepts],
                "lemma": ["science", "learn"],
            }
            if query.term and query.term.casefold() not in {
                t.casefold()
                for t in p.keywords + [c.name for c in p.concepts] + values["lemma"]
            }:
                continue
            weight = p.metrics.citations if query.metric == "citations" else 1
            for key in product(*(values[d] for d in dims)):
                counts[key] += weight
        return AggregateResult(
            metric=query.metric,
            group_by=dims,
            points=[
                AggregatePoint(dimensions=dict(zip(dims, k)), value=v)
                for k, v in sorted(counts.items())
            ],
            data_as_of=AS_OF,
        )

    async def top(self, query: RankingQuery) -> Ranking:
        pubs = self.selected(query)
        objects: dict[str, tuple[str, float]] = {}
        if query.entity == "publication":
            objects = {p.id: (p.title, getattr(p.metrics, query.metric)) for p in pubs}
        elif query.entity == "author":
            for p in pubs:
                for a in p.authors:
                    old = objects.get(a.id, (a.name, 0))[1]
                    objects[a.id] = (
                        a.name,
                        old
                        + (p.metrics.citations if query.metric == "citations" else 1),
                    )
        else:
            for p in pubs:
                for id, label in (
                    [(k, k) for k in p.keywords]
                    if query.entity == "keyword"
                    else [(c.id, c.name) for c in p.concepts]
                ):
                    objects[id] = (label, objects.get(id, (label, 0))[1] + 1)
        ordered = sorted(objects.items(), key=lambda x: (-x[1][1], x[0]))[: query.limit]
        return Ranking(
            metric=query.metric,
            items=[
                RankingItem(
                    position=i,
                    object=RankingObject(entity_type=query.entity, id=id, label=label),
                    value=value,
                )
                for i, (id, (label, value)) in enumerate(ordered, 1)
            ],
        )

    async def graph(self, query: GraphQuery) -> Graph:
        pubs = self.selected(query)
        nodes: dict[str, GraphNode] = {}
        edges: dict[tuple[str, str], GraphEdge] = {}
        if query.type == "citation":
            nodes = {
                p.id: GraphNode(id=p.id, type="publication", label=p.title)
                for p in pubs
            }
            for r in self.store.references:
                if r.citing_publication_id in nodes and r.cited_publication_id in nodes:
                    edge = GraphEdge(
                        source=r.citing_publication_id,
                        target=r.cited_publication_id,
                        type="cites",
                    )
                    edges[edge.source, edge.target] = edge
        else:
            for p in pubs:
                entities = p.authors if query.type == "coauthor" else p.concepts
                for e in entities:
                    nodes[e.id] = GraphNode(
                        id=e.id,
                        type="author" if query.type == "coauthor" else "concept",
                        label=e.name,
                    )
                for i, a in enumerate(entities):
                    for b in entities[i + 1 :]:
                        key = tuple(sorted([a.id, b.id]))
                        old = edges.get(key)
                        edges[key] = GraphEdge(
                            source=key[0],
                            target=key[1],
                            type="coauthor" if query.type == "coauthor" else "cooccurs",
                            weight=(old.weight if old else 0) + 1,
                        )
        reachable = list(nodes)
        if query.root_id:
            # Validate root independently of category/period filters.
            all_ids = (
                set(self.store.publications)
                if query.type == "citation"
                else set(self.store.authors)
                if query.type == "coauthor"
                else {
                    c.id for p in self.store.publications.values() for c in p.concepts
                }
            )
            if query.root_id not in all_ids:
                raise not_found("graph_root")
            reachable = []
            if query.root_id in nodes:
                queue = deque([(query.root_id, 0)])
                seen = {query.root_id}
                while queue:
                    current, depth = queue.popleft()
                    reachable.append(current)
                    if depth >= query.max_depth:
                        continue
                    for edge in edges.values():
                        neighbor = (
                            edge.target
                            if edge.source == current
                            else edge.source
                            if edge.target == current
                            else None
                        )
                        if neighbor and neighbor not in seen:
                            seen.add(neighbor)
                            queue.append((neighbor, depth + 1))
        selected = reachable[: query.max_nodes]
        ids = set(selected)
        return Graph(
            nodes=[nodes[id] for id in selected],
            edges=[e for e in edges.values() if e.source in ids and e.target in ids],
            truncated=len(selected) < len(reachable),
            data_as_of=AS_OF,
        )


class MockAnalysisJobAdapter(BaseAdapter, AnalysisJobPort):
    async def start(self, command: JobRequest) -> AnalysisJob:
        for id in command.publication_ids:
            self.publication(id)
        self.store.job_sequence += 1
        run_id = None
        if command.operation == "clustering":
            self.store.run_sequence += 1
            run_id = f"run-{self.store.run_sequence}"
            self.store.runs[run_id] = ClusteringRunDetail(
                run_id=run_id,
                algorithm=command.parameters.algorithm,
                parameters=command.parameters,
                status="pending",
                processed_publications=0,
                metrics=ClusterMetrics(clusters_count=0),
                clusters=[],
            )
        job = AnalysisJob(
            job_id=f"job-{self.store.job_sequence}",
            operation=command.operation,
            status="pending",
            progress=0,
            created_at=datetime.now(timezone.utc),
            run_id=run_id,
        )
        self.store.jobs[job.job_id] = job
        return job.model_copy(deep=True)

    async def get(self, job_id: str) -> AnalysisJob:
        if job_id not in self.store.jobs:
            raise not_found("job")
        return self.store.jobs[job_id].model_copy(deep=True)

    async def runs(self, pagination: Pagination) -> Page[ClusteringRun]:
        runs = [
            ClusteringRun.model_validate(r.model_dump(exclude={"clusters"}))
            for r in self.store.runs.values()
        ]
        return paginate(runs, pagination)

    async def run(self, run_id: str) -> ClusteringRunDetail:
        if run_id not in self.store.runs:
            raise not_found("clustering_run")
        return self.store.runs[run_id].model_copy(deep=True)
