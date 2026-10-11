import pytest

from fragmented_keys import Configuration, MemoryHandler


@pytest.fixture(autouse=True)
def _reset_config():
    """Reset global configuration between tests."""
    Configuration.reset()
    yield
    Configuration.reset()


@pytest.fixture()
def memory_handler():
    handler = MemoryHandler()
    Configuration.set_default_cache_handler(handler)
    Configuration.set_global_prefix("Test")
    return handler
