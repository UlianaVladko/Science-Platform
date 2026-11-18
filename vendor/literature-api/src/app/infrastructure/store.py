"""One independent, in-memory fixture store per application instance."""

import json
from datetime import datetime, timezone
from importlib.resources import files
from app.domain.models import (
    Author,
    Category,
    Publication,
    Reference,
    AnalysisJob,
    ClusteringRunDetail,
    ClusteringParameters,
    ClusterMetrics,
    Cluster,
    PublicationFile,
)

AS_OF = datetime(2026, 11, 1, tzinfo=timezone.utc)
FIXTURES = files("app.infrastructure").joinpath("fixtures")


class MockStore:
    def __init__(self) -> None:
        data = json.loads(FIXTURES.joinpath("corpus.json").read_text())
        self.authors = {a["id"]: Author.model_validate(a) for a in data["authors"]}
        self.categories = {
            c["id"]: Category.model_validate(c) for c in data["categories"]
        }
        self.publications = {
            p["id"]: Publication.model_validate(p) for p in data["publications"]
        }
        self.references = [Reference.model_validate(r) for r in data["references"]]
        self.files: dict[tuple[str, str], bytes] = {}
        for p in self.publications.values():
            for kind, filename, mime in [
                ("raw", "sample.pdf", "application/pdf"),
                ("markdown", "sample.md", "text/markdown"),
            ]:
                content = FIXTURES.joinpath(filename).read_bytes()
                self.files[p.id, kind] = content
                p.files.append(
                    PublicationFile(
                        kind=kind, content_type=mime, size_bytes=len(content)
                    )
                )
        self.jobs = {
            f"job-{status}": AnalysisJob(
                job_id=f"job-{status}",
                operation="extract_terms",
                status=status,
                progress=100
                if status == "completed"
                else 45
                if status == "running"
                else 0,
                created_at=AS_OF,
                completed_at=AS_OF if status in {"completed", "failed"} else None,
                result={"processed_publications": 2} if status == "completed" else None,
                error={"code": "DEMO_FAILURE", "message": "Demonstration failure"}
                if status == "failed"
                else None,
            )
            for status in ["pending", "running", "completed", "failed"]
        }
        self.runs = {
            "run-1": ClusteringRunDetail(
                run_id="run-1",
                algorithm="kmeans",
                parameters=ClusteringParameters(),
                status="completed",
                processed_publications=4,
                metrics=ClusterMetrics(silhouette=0.61, clusters_count=2),
                clusters=[
                    Cluster(
                        cluster_id="cluster-1",
                        size=2,
                        publication_ids=["pub-1", "pub-3"],
                        keywords=["text"],
                    ),
                    Cluster(
                        cluster_id="cluster-2",
                        size=2,
                        publication_ids=["pub-2", "pub-4"],
                        keywords=["learning"],
                    ),
                ],
            )
        }
        self.publication_sequence = 4
        self.job_sequence = 0
        self.run_sequence = 1
