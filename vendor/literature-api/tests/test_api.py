import ast
from pathlib import Path
import pytest
from pydantic import TypeAdapter
from app.api.dependencies import get_publication_service
from app.application.services import PublicationService
from app.domain.models import Publication, PublicationType
from app.schemas.errors import ErrorResponse

BASE = "/api/v1"
CREATE = {
    "title": "A new publication",
    "publication_type": "article",
    "authors": [{"author_id": "author-1"}],
    "category_ids": ["cat-nlp"],
}
OPERATIONS = [
    ("GET", "/categories", {}, 200),
    ("POST", "/search", {"json": {"query": "text", "include_facets": True}}, 200),
    ("GET", "/search/suggest?q=Neural", {}, 200),
    ("GET", "/publications/pub-1", {}, 200),
    ("POST", "/publications", {"json": CREATE}, 201),
    ("PATCH", "/publications/pub-1", {"json": {"title": "Updated"}}, 200),
    ("DELETE", "/publications/pub-1", {}, 204),
    (
        "PUT",
        "/publications/pub-1/files/raw",
        {"content": b"%PDF-1.4\nDemo", "headers": {"Content-Type": "application/pdf"}},
        200,
    ),
    ("GET", "/publications/pub-1/files/raw", {}, 200),
    ("GET", "/publications/pub-3/references", {}, 200),
    ("POST", "/publications/pub-1/view", {}, 200),
    ("GET", "/authors/author-1", {}, 200),
    ("GET", "/analytics/summary", {}, 200),
    (
        "GET",
        "/analytics/aggregates?metric=publications&group_by=publication_year",
        {},
        200,
    ),
    ("GET", "/analytics/top", {}, 200),
    ("GET", "/graphs", {}, 200),
    (
        "POST",
        "/analysis/jobs",
        {"json": {"operation": "clustering", "publication_ids": ["pub-1", "pub-2"]}},
        202,
    ),
    ("GET", "/jobs/job-completed", {}, 200),
    ("GET", "/analysis/clustering/runs", {}, 200),
    ("GET", "/analysis/clustering/runs/run-1", {}, 200),
]


@pytest.mark.parametrize("method,path,kwargs,status", OPERATIONS)
def test_all_operations_and_response_models(app, client, method, path, kwargs, status):
    response = client.request(method, BASE + path, **kwargs)
    assert response.status_code == status, response.text
    assert response.headers["x-request-id"].startswith("req-")
    if status == 204:
        assert response.content == b""
    elif "application/json" in response.headers.get("content-type", ""):
        from app.domain import models as m

        models = [
            list[m.Category],
            m.SearchResponse,
            m.Suggestions,
            m.Publication,
            m.Publication,
            m.Publication,
            None,
            m.FileAccepted,
            None,
            m.Page[m.Reference],
            m.ViewCount,
            m.Author,
            m.CorpusSummary,
            m.AggregateResult,
            m.Ranking,
            m.Graph,
            m.AnalysisJob,
            m.AnalysisJob,
            m.Page[m.ClusteringRun],
            m.ClusteringRunDetail,
        ]
        index = next(
            i for i, x in enumerate(OPERATIONS) if x[0] == method and x[1] == path
        )
        TypeAdapter(models[index]).validate_python(response.json())


def test_openapi_exact_operations_and_documentation(client):
    response = client.get("/openapi.json")
    assert response.status_code == 200
    schema = response.json()
    operations = [
        (p, m, o)
        for p, item in schema["paths"].items()
        for m, o in item.items()
        if m in {"get", "post", "patch", "put", "delete"}
    ]
    assert len(operations) == 20
    expected = {("/api/v1" + p.split("?")[0], m.lower()) for m, p, _, _ in OPERATIONS}
    normalize = lambda p: (
        p.replace("pub-1", "{id}")
        .replace("pub-3", "{id}")
        .replace("author-1", "{id}")
        .replace("job-completed", "{job_id}")
        .replace("run-1", "{run_id}")
        .replace("/files/raw", "/files/{kind}")
    )
    assert {(p, m) for p, m, _ in operations} == {
        (normalize(p), m) for p, m in expected
    }
    for path, method, operation in operations:
        assert operation["summary"] and operation["description"] and operation["tags"]
        for status in ["400", "404", "409", "422", "500"]:
            ref = operation["responses"][status]["content"]["application/json"][
                "schema"
            ]["$ref"]
            assert ref.endswith("/ErrorResponse")
    assert client.get("/docs").status_code == 200
    assert client.get("/redoc").status_code == 200
    assert (
        "discriminator"
        in schema["components"]["schemas"]["SearchResponse"]["properties"]["items"][
            "items"
        ]
    )


@pytest.mark.parametrize(
    "path",
    [
        "/search/suggest?q=x&limit=0",
        "/analysis/clustering/runs?page=0",
        "/analytics/aggregates?metric=sql",
        "/analytics/aggregates?metric=citations&group_by=lemma",
        "/analytics/aggregates?metric=term_frequency&group_by=publication_year",
        "/analytics/aggregates?group_by=category,category",
        "/analytics/top?entity=author&metric=views",
        "/graphs?type=invalid",
        "/graphs?max_nodes=0",
        "/graphs?max_depth=6",
        "/publications/pub-1/references?direction=both",
        "/categories?include_counts=maybe",
    ],
)
def test_invalid_query(client, path):
    r = client.get(BASE + path)
    assert r.status_code == 422, r.text
    e = ErrorResponse.model_validate(r.json())
    assert e.error.code == "VALIDATION_ERROR"
    assert e.error.request_id == r.headers["x-request-id"]


@pytest.mark.parametrize(
    "body",
    [
        {"mode": "similar"},
        {
            "entity_type": "author",
            "mode": "similar",
            "similar_to_publication_id": "pub-1",
        },
        {"filters": {"year_from": 2026, "year_to": 2020}},
        {"page_size": 101},
        {"filters": {"typo": 1}},
        {"similar_to_publication_id": "pub-1"},
        {"entity_type": "author", "sort": "views"},
    ],
)
def test_invalid_search(client, body):
    assert client.post(BASE + "/search", json=body).status_code == 422


@pytest.mark.parametrize(
    "path",
    [
        "/publications/nope",
        "/authors/nope",
        "/jobs/nope",
        "/analysis/clustering/runs/nope",
        "/publications/nope/files/raw",
        "/publications/nope/references",
        "/categories?parent_id=nope",
        "/graphs?root_id=nope",
    ],
)
def test_missing(client, path):
    r = client.get(BASE + path)
    assert r.status_code == 404
    assert ErrorResponse.model_validate(r.json()).error.code.endswith("_NOT_FOUND")


def test_mutations_and_atomic_conflicts(client):
    body = {
        **CREATE,
        "external_identifiers": [{"scheme": "doi", "value": "10.demo/new"}],
    }
    r = client.post(BASE + "/publications", json=body)
    id = r.json()["id"]
    assert client.get(BASE + "/publications/" + id).json()["title"] == CREATE["title"]
    assert client.post(BASE + "/publications", json=body).status_code == 409
    conflict = client.request(
        "PATCH",
        BASE + "/publications/" + id,
        json={
            "title": "Bad",
            "external_identifiers": [{"scheme": "doi", "value": "10.0000/mock-1"}],
        },
    )
    assert conflict.status_code == 409
    assert client.get(BASE + "/publications/" + id).json()["title"] == CREATE["title"]
    for patch in [
        {},
        {"id": "overwrite"},
        {"title": None},
        {"authors": None},
        {"metrics": {}},
    ]:
        assert (
            client.request(
                "PATCH", BASE + "/publications/" + id, json=patch
            ).status_code
            == 422
        )
    assert (
        client.request(
            "PATCH",
            BASE + "/publications/" + id,
            json={"abstract": None, "publisher": "Demo"},
        ).status_code
        == 200
    )
    a = client.post(BASE + "/publications/" + id + "/view").json()["views"]
    assert client.post(BASE + "/publications/" + id + "/view").json()["views"] == a + 1
    assert client.request("DELETE", BASE + "/publications/" + id).status_code == 204
    assert client.get(BASE + "/publications/" + id).status_code == 404


def test_pagination_filters_and_empty(client):
    a = client.post(BASE + "/search", json={"page": 1, "page_size": 2}).json()
    b = client.post(BASE + "/search", json={"page": 2, "page_size": 2}).json()
    assert a["total"] == b["total"] == 4
    assert len(a["items"]) == len(b["items"]) == 2
    assert {x["entity"]["id"] for x in a["items"]}.isdisjoint(
        x["entity"]["id"] for x in b["items"]
    )
    assert client.post(BASE + "/search", json={"page": 99}).json()["items"] == []
    assert (
        client.post(BASE + "/search", json={"query": "absentword"}).json()["total"] == 0
    )
    for filters, count in [
        ({"author": "author-1"}, 2),
        ({"organization": "org-1"}, 4),
        ({"publication_type": ["book"]}, 1),
        ({"category": "cat-cs"}, 4),
        ({"language": "ru"}, 0),
        ({"journal": "Science Demo"}, 1),
        ({"terms": ["learning"]}, 2),
        ({"year_from": 2026, "year_to": 2026}, 1),
    ]:
        assert (
            client.post(BASE + "/search", json={"filters": filters}).json()["total"]
            == count
        )
    similar = client.post(
        BASE + "/search", json={"mode": "similar", "similar_to_publication_id": "pub-1"}
    ).json()
    assert all(x["entity"]["id"] != "pub-1" for x in similar["items"])
    authors = client.post(
        BASE + "/search", json={"entity_type": "author", "query": "Anna"}
    ).json()
    assert authors["items"][0]["entity_type"] == "author"


def test_files(client):
    r = client.get(BASE + "/publications/pub-1/files/raw")
    assert r.headers["content-type"] == "application/pdf" and r.content.startswith(
        b"%PDF-"
    )
    assert b"%%EOF" in r.content
    path = BASE + "/publications/pub-1/files/markdown"
    assert client.get(path).headers["content-type"].startswith("text/markdown")
    assert (
        client.request(
            "PUT",
            path,
            content=b"# Updated",
            headers={"Content-Type": "text/markdown; charset=utf-8"},
        ).status_code
        == 200
    )
    assert client.get(path).content == b"# Updated"
    for content, mime in [(b"xx", "application/pdf"), (b"\xff", "text/markdown")]:
        assert (
            client.request(
                "PUT", path, content=content, headers={"Content-Type": mime}
            ).status_code
            == 400
        )
    assert (
        client.request(
            "PUT", path, content=b"", headers={"Content-Type": "text/markdown"}
        ).status_code
        == 422
    )
    assert client.get(BASE + "/publications/pub-1/files/other").status_code == 422


def test_consistency_references_categories_graphs(client):
    card = client.get(BASE + "/publications/pub-1").json()
    for a in card["authors"]:
        assert client.get(BASE + "/authors/" + a["id"]).json()["name"] == a["name"]
    refs = client.get(BASE + "/publications/pub-3/references").json()
    assert any(
        r["cited_publication_id"] is None and r["raw_reference"] for r in refs["items"]
    )
    incoming = client.get(
        BASE + "/publications/pub-1/references?direction=incoming"
    ).json()
    assert incoming["items"][0]["publication"]["id"] == "pub-3"
    tree = client.get(BASE + "/categories").json()
    assert tree[0]["publications_count"] == 4 and len(tree[0]["children"]) == 2
    assert (
        client.get(BASE + "/categories?include_counts=false").json()[0][
            "publications_count"
        ]
        is None
    )
    for typ in ["citation", "coauthor", "concept"]:
        graph = client.get(BASE + "/graphs?type=" + typ + "&max_nodes=1").json()
        assert len(graph["nodes"]) <= 1 and graph["truncated"]
        ids = {n["id"] for n in graph["nodes"]}
        assert all(e["source"] in ids and e["target"] in ids for e in graph["edges"])
    assert (
        len(client.get(BASE + "/graphs?root_id=pub-1&max_depth=0").json()["nodes"]) == 1
    )


@pytest.mark.parametrize("status", ["pending", "running", "completed", "failed"])
def test_job_states(client, status):
    job = client.get(BASE + "/jobs/job-" + status).json()
    assert job["status"] == status
    assert (job["error"] is not None) == (status == "failed")
    assert (job["result"] is not None) == (status == "completed")


def test_job_creation_and_runs(client):
    for operation in ["process_publications", "extract_terms", "clustering"]:
        response = client.post(
            BASE + "/analysis/jobs",
            json={"operation": operation, "publication_ids": ["pub-1", "pub-2"]},
        )
        assert response.status_code == 202, response.text
        job = response.json()
        assert client.get(BASE + "/jobs/" + job["job_id"]).json()["status"] == "pending"
        if operation == "clustering":
            assert (
                client.get(BASE + "/analysis/clustering/runs/" + job["run_id"]).json()[
                    "status"
                ]
                == "pending"
            )
    assert (
        client.get(BASE + "/analysis/clustering/runs?page_size=1").json()["total"] == 2
    )
    assert (
        client.post(
            BASE + "/analysis/jobs",
            json={"operation": "clustering", "publication_ids": ["missing"]},
        ).status_code
        == 404
    )
    assert (
        client.post(
            BASE + "/analysis/jobs",
            json={
                "operation": "clustering",
                "publication_ids": ["pub-1"],
                "parameters": {"arbitrary": 42},
            },
        ).status_code
        == 422
    )
    run = client.get(BASE + "/analysis/clustering/runs/run-1").json()
    assert sum(c["size"] for c in run["clusters"]) == run["processed_publications"]


def test_overrides_and_server_errors(app, client):
    class Port:
        async def get(self, publication_id: str) -> Publication:
            return Publication(
                id=publication_id,
                title="Overridden port",
                publication_type=PublicationType.book,
                language="en",
                authors=[],
            )

    app.dependency_overrides[get_publication_service] = lambda: PublicationService(
        Port()
    )
    assert (
        client.get(BASE + "/publications/anything").json()["title"] == "Overridden port"
    )

    class BrokenPort:
        async def get(self, publication_id: str) -> Publication:
            raise RuntimeError("SECRET_INTERNAL_DETAIL")

    app.dependency_overrides[get_publication_service] = lambda: PublicationService(
        BrokenPort()
    )
    r = client.get(BASE + "/publications/anything")
    assert r.status_code == 500
    assert "SECRET" not in r.text
    assert ErrorResponse.model_validate(r.json()).error.code == "INTERNAL_ERROR"
    app.dependency_overrides.clear()


def test_layer_dependencies():
    root = Path(__file__).parents[1] / "src/app"
    for folder in ["domain", "application", "schemas"]:
        for file in (root / folder).rglob("*.py"):
            tree = ast.parse(file.read_text())
            for node in ast.walk(tree):
                if isinstance(node, ast.ImportFrom):
                    assert not (node.module or "").startswith(
                        ("fastapi", "starlette", "app.infrastructure", "app.api")
                    )
    for file in (root / "api/routers").glob("*.py"):
        assert "app.infrastructure" not in file.read_text()


@pytest.mark.parametrize(
    "metric,group_by,extra",
    [
        ("publications", "publication_year", {}),
        ("added_documents", "created_date", {}),
        ("publications", "category", {}),
        ("publications", "language", {}),
        ("publications", "publication_type", {}),
        ("publications", "organization", {}),
        ("publications", "status", {}),
        ("publications", "source", {}),
        ("publications", "coauthors_count", {}),
        ("citations", "citation_year", {}),
        ("term_frequency", "keyword", {}),
        ("term_frequency", "concept", {}),
        ("term_frequency", "lemma", {}),
        ("term_frequency", "publication_year", {"term": "Neural networks"}),
        ("publications", "publication_year,category", {}),
    ],
)
def test_aggregate_scenarios(client, metric, group_by, extra):
    from app.domain.models import AggregateResult

    response = client.get(
        BASE + "/analytics/aggregates",
        params={"metric": metric, "group_by": group_by, **extra},
    )
    assert response.status_code == 200, response.text
    result = AggregateResult.model_validate(response.json())
    assert result.group_by == group_by.split(",")
    assert result.points
    assert all(set(p.dimensions) == set(result.group_by) for p in result.points)


@pytest.mark.parametrize(
    "entity,metric",
    [
        ("publication", "citations"),
        ("publication", "views"),
        ("publication", "downloads"),
        ("author", "citations"),
        ("author", "publications"),
        ("keyword", "frequency"),
        ("concept", "frequency"),
    ],
)
def test_rankings(client, entity, metric):
    data = client.get(
        BASE + "/analytics/top", params={"entity": entity, "metric": metric, "limit": 2}
    ).json()
    assert len(data["items"]) <= 2
    assert [i["position"] for i in data["items"]] == list(
        range(1, len(data["items"]) + 1)
    )
    assert all(i["object"]["entity_type"] == entity for i in data["items"])
    values = [i["value"] for i in data["items"]]
    assert values == sorted(values, reverse=True)


def test_author_order_affiliations_and_missing_links(client):
    body = {
        **CREATE,
        "authors": [
            {"author_id": "author-3", "affiliation_ids": []},
            {"author_id": "author-1"},
        ],
    }
    response = client.post(BASE + "/publications", json=body)
    assert response.status_code == 201, response.text
    authors = response.json()["authors"]
    assert [a["id"] for a in authors] == ["author-3", "author-1"]
    assert authors[0]["affiliations"] == []
    for field, value in [
        ("authors", [{"author_id": "missing"}]),
        ("category_ids", ["missing"]),
    ]:
        assert (
            client.post(
                BASE + "/publications", json={**CREATE, field: value}
            ).status_code
            == 404
        )
    assert (
        client.post(
            BASE + "/publications",
            json={
                **CREATE,
                "authors": [{"author_id": "author-1", "affiliation_ids": ["missing"]}],
            },
        ).status_code
        == 404
    )
    assert (
        client.post(
            BASE + "/publications",
            json={
                **CREATE,
                "authors": [{"author_id": "author-1"}, {"author_id": "author-1"}],
            },
        ).status_code
        == 400
    )


def test_mock_instances_and_returned_models_are_isolated(app):
    import asyncio
    from app.main import create_app

    async def check():
        first = app.state.services.publication
        before = await first.get("pub-1")
        before.title = "Changed by caller"
        assert (await first.get("pub-1")).title != "Changed by caller"
        await first.view("pub-1")
        fresh = await create_app().state.services.publication.get("pub-1")
        assert (await first.get("pub-1")).metrics.views == fresh.metrics.views + 1

    asyncio.run(check())
