# Live Redis tests are tagged `:redis` and excluded by default so the suite
# runs without a broker. Enable them with:
#
#     mix test --include redis
#
# Point at a specific server via FK_REDIS_HOST / FK_REDIS_PORT (defaults
# 127.0.0.1:6379). For the bundled docker-compose service use:
#
#     FK_REDIS_PORT=6380 mix test --include redis
ExUnit.start(exclude: [:redis])
