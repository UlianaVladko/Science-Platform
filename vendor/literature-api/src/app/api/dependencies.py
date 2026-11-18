"""Composition root and FastAPI dependency providers."""

from dataclasses import dataclass
from fastapi import Request
from app.infrastructure.store import MockStore
from app.infrastructure.adapters import (
    MockPublicationAdapter,
    MockSearchAdapter,
    MockCategoryAdapter,
    MockAuthorAdapter,
    MockAnalyticsAdapter,
    MockAnalysisJobAdapter,
)
from app.application.services import (
    PublicationService,
    SearchService,
    CategoryService,
    AuthorService,
    AnalyticsService,
    AnalysisJobService,
)


@dataclass
class Services:
    publication: PublicationService
    search: SearchService
    category: CategoryService
    author: AuthorService
    analytics: AnalyticsService
    analysisjob: AnalysisJobService


def build_services() -> Services:
    store = MockStore()
    return Services(
        publication=PublicationService(MockPublicationAdapter(store)),
        search=SearchService(MockSearchAdapter(store)),
        category=CategoryService(MockCategoryAdapter(store)),
        author=AuthorService(MockAuthorAdapter(store)),
        analytics=AnalyticsService(MockAnalyticsAdapter(store)),
        analysisjob=AnalysisJobService(MockAnalysisJobAdapter(store)),
    )


def get_publication_service(request: Request) -> PublicationService:
    return request.app.state.services.publication


def get_search_service(request: Request) -> SearchService:
    return request.app.state.services.search


def get_category_service(request: Request) -> CategoryService:
    return request.app.state.services.category


def get_author_service(request: Request) -> AuthorService:
    return request.app.state.services.author


def get_analytics_service(request: Request) -> AnalyticsService:
    return request.app.state.services.analytics


def get_analysisjob_service(request: Request) -> AnalysisJobService:
    return request.app.state.services.analysisjob
