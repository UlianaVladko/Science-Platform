"""Static OpenAPI examples, separate from runtime fixture implementations."""

from typing import Any

EXAMPLES: dict[str, Any] = {
    "categories": [
        {
            "id": "cat-cs",
            "slug": "computer-science",
            "name": "Computer science",
            "description": "Computational methods",
            "parent_id": None,
            "publications_count": 4,
            "children": [
                {
                    "id": "cat-nlp",
                    "slug": "nlp",
                    "name": "Natural language processing",
                    "description": "Scientific text analysis",
                    "parent_id": "cat-cs",
                    "publications_count": 2,
                    "children": [],
                },
                {
                    "id": "cat-ml",
                    "slug": "machine-learning",
                    "name": "Machine learning",
                    "description": "Learning from data",
                    "parent_id": "cat-cs",
                    "publications_count": 2,
                    "children": [],
                },
            ],
        }
    ],
    "search": {
        "items": [
            {
                "entity_type": "publication",
                "score": 1.0,
                "entity": {
                    "id": "pub-1",
                    "title": "Neural scientific text analysis",
                    "abstract": "A demonstration scientific publication about text "
                    "and learning.",
                    "publication_type": "article",
                    "publication_date": "2025-02-01",
                    "language": "en",
                    "authors": [
                        {
                            "id": "author-1",
                            "name": "Anna Petrova",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                        {
                            "id": "author-2",
                            "name": "Ivan Smirnov",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                    ],
                    "metrics": {"citations": 1, "views": 10, "downloads": 2},
                },
                "snippets": [
                    "A demonstration scientific publication about text and learning."
                ],
            },
            {
                "entity_type": "publication",
                "score": 0.5,
                "entity": {
                    "id": "pub-2",
                    "title": "Foundations of machine learning",
                    "abstract": "A demonstration scientific publication about text "
                    "and learning.",
                    "publication_type": "book",
                    "publication_date": "2026-02-01",
                    "language": "en",
                    "authors": [
                        {
                            "id": "author-2",
                            "name": "Ivan Smirnov",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                        {
                            "id": "author-3",
                            "name": "Maria Kim",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                    ],
                    "metrics": {"citations": 1, "views": 20, "downloads": 4},
                },
                "snippets": [
                    "A demonstration scientific publication about text and learning."
                ],
            },
            {
                "entity_type": "publication",
                "score": 0.3333333333333333,
                "entity": {
                    "id": "pub-3",
                    "title": "Semantic search for science",
                    "abstract": "A demonstration scientific publication about text "
                    "and learning.",
                    "publication_type": "preprint",
                    "publication_date": "2024-02-01",
                    "language": "en",
                    "authors": [
                        {
                            "id": "author-1",
                            "name": "Anna Petrova",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                        {
                            "id": "author-3",
                            "name": "Maria Kim",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                    ],
                    "metrics": {"citations": 0, "views": 30, "downloads": 6},
                },
                "snippets": [
                    "A demonstration scientific publication about text and learning."
                ],
            },
            {
                "entity_type": "publication",
                "score": 0.25,
                "entity": {
                    "id": "pub-4",
                    "title": "Scientific knowledge networks",
                    "abstract": "A demonstration scientific publication about text "
                    "and learning.",
                    "publication_type": "monograph",
                    "publication_date": "2025-02-01",
                    "language": "en",
                    "authors": [
                        {
                            "id": "author-3",
                            "name": "Maria Kim",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        }
                    ],
                    "metrics": {"citations": 0, "views": 40, "downloads": 8},
                },
                "snippets": [
                    "A demonstration scientific publication about text and learning."
                ],
            },
        ],
        "total": 4,
        "page": 1,
        "page_size": 20,
        "facets": [{"field": "language", "buckets": [{"value": "en", "count": 4}]}],
    },
    "suggest": {
        "items": [
            {
                "type": "publication",
                "text": "Neural scientific text analysis",
                "matches_count": 1,
            },
            {"type": "term", "text": "Neural networks", "matches_count": 1},
        ]
    },
    "get_publication": {
        "id": "pub-1",
        "title": "Neural scientific text analysis",
        "abstract": "A demonstration scientific publication about text and learning.",
        "publication_type": "article",
        "publication_date": "2025-02-01",
        "language": "en",
        "authors": [
            {
                "id": "author-1",
                "name": "Anna Petrova",
                "orcid": None,
                "affiliations": [
                    {"id": "org-1", "name": "Institute of Scientific Computing"}
                ],
            },
            {
                "id": "author-2",
                "name": "Ivan Smirnov",
                "orcid": None,
                "affiliations": [
                    {"id": "org-1", "name": "Institute of Scientific Computing"}
                ],
            },
        ],
        "metrics": {"citations": 1, "views": 10, "downloads": 2},
        "external_identifiers": [{"scheme": "doi", "value": "10.0000/mock-1"}],
        "journal": "Science Demo",
        "publisher": None,
        "categories": [
            {
                "id": "cat-nlp",
                "slug": "nlp",
                "name": "Natural language processing",
                "description": "Scientific text analysis",
                "parent_id": "cat-cs",
                "publications_count": None,
                "children": [],
            }
        ],
        "keywords": ["science", "text"],
        "concepts": [
            {"id": "concept-1", "name": "Neural networks"},
            {"id": "concept-2", "name": "Semantic similarity"},
        ],
        "files": [
            {"kind": "raw", "content_type": "application/pdf", "size_bytes": 601},
            {"kind": "markdown", "content_type": "text/markdown", "size_bytes": 66},
        ],
        "processing_status": "completed",
    },
    "create_publication": {
        "id": "pub-5",
        "title": "A new publication",
        "abstract": None,
        "publication_type": "article",
        "publication_date": None,
        "language": "en",
        "authors": [
            {
                "id": "author-1",
                "name": "Anna Petrova",
                "orcid": None,
                "affiliations": [
                    {"id": "org-1", "name": "Institute of Scientific Computing"}
                ],
            }
        ],
        "metrics": {"citations": 0, "views": 0, "downloads": 0},
        "external_identifiers": [],
        "journal": None,
        "publisher": None,
        "categories": [
            {
                "id": "cat-nlp",
                "slug": "nlp",
                "name": "Natural language processing",
                "description": "Scientific text analysis",
                "parent_id": "cat-cs",
                "publications_count": None,
                "children": [],
            }
        ],
        "keywords": [],
        "concepts": [],
        "files": [],
        "processing_status": "pending",
    },
    "update_publication": {
        "id": "pub-1",
        "title": "Updated",
        "abstract": "A demonstration scientific publication about text and learning.",
        "publication_type": "article",
        "publication_date": "2025-02-01",
        "language": "en",
        "authors": [
            {
                "id": "author-1",
                "name": "Anna Petrova",
                "orcid": None,
                "affiliations": [
                    {"id": "org-1", "name": "Institute of Scientific Computing"}
                ],
            },
            {
                "id": "author-2",
                "name": "Ivan Smirnov",
                "orcid": None,
                "affiliations": [
                    {"id": "org-1", "name": "Institute of Scientific Computing"}
                ],
            },
        ],
        "metrics": {"citations": 1, "views": 10, "downloads": 2},
        "external_identifiers": [{"scheme": "doi", "value": "10.0000/mock-1"}],
        "journal": "Science Demo",
        "publisher": None,
        "categories": [
            {
                "id": "cat-nlp",
                "slug": "nlp",
                "name": "Natural language processing",
                "description": "Scientific text analysis",
                "parent_id": "cat-cs",
                "publications_count": None,
                "children": [],
            }
        ],
        "keywords": ["science", "text"],
        "concepts": [
            {"id": "concept-1", "name": "Neural networks"},
            {"id": "concept-2", "name": "Semantic similarity"},
        ],
        "files": [
            {"kind": "raw", "content_type": "application/pdf", "size_bytes": 601},
            {"kind": "markdown", "content_type": "text/markdown", "size_bytes": 66},
        ],
        "processing_status": "completed",
    },
    "put_file": {
        "publication_id": "pub-1",
        "file": {"kind": "raw", "content_type": "application/pdf", "size_bytes": 13},
        "accepted": True,
    },
    "references": {
        "items": [
            {
                "citing_publication_id": "pub-3",
                "cited_publication_id": "pub-1",
                "raw_reference": "Petrova and Smirnov. Neural scientific text analysis.",
                "publication": {
                    "id": "pub-1",
                    "title": "Neural scientific text analysis",
                    "abstract": "A demonstration scientific publication "
                    "about text and learning.",
                    "publication_type": "article",
                    "publication_date": "2025-02-01",
                    "language": "en",
                    "authors": [
                        {
                            "id": "author-1",
                            "name": "Anna Petrova",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                        {
                            "id": "author-2",
                            "name": "Ivan Smirnov",
                            "orcid": None,
                            "affiliations": [
                                {
                                    "id": "org-1",
                                    "name": "Institute of Scientific Computing",
                                }
                            ],
                        },
                    ],
                    "metrics": {"citations": 1, "views": 10, "downloads": 2},
                },
            },
            {
                "citing_publication_id": "pub-3",
                "cited_publication_id": None,
                "raw_reference": "An external work not in the corpus (2020).",
                "publication": None,
            },
        ],
        "total": 2,
        "page": 1,
        "page_size": 20,
    },
    "view": {"publication_id": "pub-1", "views": 11},
    "author": {
        "id": "author-1",
        "name": "Anna Petrova",
        "orcid": None,
        "affiliations": [{"id": "org-1", "name": "Institute of Scientific Computing"}],
        "categories": [
            {
                "id": "cat-nlp",
                "slug": "nlp",
                "name": "Natural language processing",
                "description": "Scientific text analysis",
                "parent_id": "cat-cs",
                "publications_count": None,
                "children": [],
            }
        ],
        "publications_count": 2,
        "citations_count": 1,
        "h_index": 1,
    },
    "summary": {
        "total_publications": 4,
        "total_authors": 3,
        "total_citations": 2,
        "total_downloads": 20,
        "avg_citations_per_publication": 0.5,
        "publications_this_period": 1,
        "new_authors_this_period": 1,
    },
    "aggregates": {
        "metric": "publications",
        "group_by": ["publication_year"],
        "points": [
            {"dimensions": {"publication_year": "2024"}, "value": 1.0},
            {"dimensions": {"publication_year": "2025"}, "value": 2.0},
            {"dimensions": {"publication_year": "2026"}, "value": 1.0},
        ],
        "data_as_of": "2026-11-01T00:00:00Z",
    },
    "top": {
        "metric": "citations",
        "items": [
            {
                "position": 1,
                "object": {
                    "entity_type": "publication",
                    "id": "pub-1",
                    "label": "Neural scientific text analysis",
                },
                "value": 1.0,
            },
            {
                "position": 2,
                "object": {
                    "entity_type": "publication",
                    "id": "pub-2",
                    "label": "Foundations of machine learning",
                },
                "value": 1.0,
            },
            {
                "position": 3,
                "object": {
                    "entity_type": "publication",
                    "id": "pub-3",
                    "label": "Semantic search for science",
                },
                "value": 0.0,
            },
            {
                "position": 4,
                "object": {
                    "entity_type": "publication",
                    "id": "pub-4",
                    "label": "Scientific knowledge networks",
                },
                "value": 0.0,
            },
        ],
    },
    "graph": {
        "nodes": [
            {
                "id": "pub-1",
                "type": "publication",
                "label": "Neural scientific text analysis",
            },
            {
                "id": "pub-2",
                "type": "publication",
                "label": "Foundations of machine learning",
            },
            {
                "id": "pub-3",
                "type": "publication",
                "label": "Semantic search for science",
            },
            {
                "id": "pub-4",
                "type": "publication",
                "label": "Scientific knowledge networks",
            },
        ],
        "edges": [
            {"source": "pub-1", "target": "pub-2", "type": "cites", "weight": None},
            {"source": "pub-3", "target": "pub-1", "type": "cites", "weight": None},
        ],
        "truncated": False,
        "data_as_of": "2026-11-01T00:00:00Z",
    },
    "start_job": {
        "job_id": "job-1",
        "operation": "clustering",
        "status": "pending",
        "progress": 0.0,
        "created_at": "2026-11-01T00:00:00Z",
        "completed_at": None,
        "result": None,
        "error": None,
        "run_id": "run-2",
    },
    "job": {
        "job_id": "job-completed",
        "operation": "extract_terms",
        "status": "completed",
        "progress": 100.0,
        "created_at": "2026-11-01T00:00:00Z",
        "completed_at": "2026-11-01T00:00:00Z",
        "result": {"processed_publications": 2, "run_id": None},
        "error": None,
        "run_id": None,
    },
    "runs": {
        "items": [
            {
                "run_id": "run-1",
                "algorithm": "kmeans",
                "parameters": {
                    "algorithm": "kmeans",
                    "feature_source": "embeddings",
                    "n_clusters": 2,
                },
                "status": "completed",
                "processed_publications": 4,
                "metrics": {"silhouette": 0.61, "clusters_count": 2},
            }
        ],
        "total": 1,
        "page": 1,
        "page_size": 20,
    },
    "run": {
        "run_id": "run-1",
        "algorithm": "kmeans",
        "parameters": {
            "algorithm": "kmeans",
            "feature_source": "embeddings",
            "n_clusters": 2,
        },
        "status": "completed",
        "processed_publications": 4,
        "metrics": {"silhouette": 0.61, "clusters_count": 2},
        "clusters": [
            {
                "cluster_id": "cluster-1",
                "size": 2,
                "publication_ids": ["pub-1", "pub-3"],
                "keywords": ["text"],
            },
            {
                "cluster_id": "cluster-2",
                "size": 2,
                "publication_ids": ["pub-2", "pub-4"],
                "keywords": ["learning"],
            },
        ],
    },
}


def success_response(name: str, status: int = 200) -> dict[int, dict[str, Any]]:
    return {
        status: {
            "description": "Successful response",
            "content": {"application/json": {"example": EXAMPLES[name]}},
        }
    }
