# [2.0.0](https://github.com/asaidimu/tasker/compare/v1.2.0...v2.0.0) (2025-06-30)


* feat(api)!: integrate context into task functions and add async submission methods ([fef048f](https://github.com/asaidimu/tasker/commit/fef048fb2471c289802930891e5f4b4ba9635046))


### BREAKING CHANGES

* The signature of task functions passed to QueueTask, RunTask, and other submission methods has changed from func(resource R) (result E, err error) to func(ctx context.Context, resource R) (result E, err error). All task implementations must be updated to accept the new context.Context parameter. This allows tasks to be cancelled during manager shutdown or when their parent context is cancelled.

- Introduce new non-blocking task submission methods: QueueTaskWithCallback, QueueTaskAsync, and their priority/once variants.
- Revamp and restructure project documentation for improved navigation and clarity, including a dedicated reference section and updated examples for v2 API.
- Implement GitHub Actions workflow for automated documentation deployment.
- Add utility script for Go module major version bumping.
- Remove deprecated documentation and project management files.

# [1.2.0](https://github.com/asaidimu/tasker/compare/v1.1.0...v1.2.0) (2025-06-30)


### Features

* **manager:** introduce comprehensive metrics, logging, and rate-based scaling ([1d6a918](https://github.com/asaidimu/tasker/commit/1d6a91843268dbb99e38013e0ea2a795a0b82075))

# [1.1.0](https://github.com/asaidimu/tasker/compare/v1.0.0...v1.1.0) (2025-06-30)


### Features

* **core:** enhance task queuing and introduce immediate shutdown ([c819e84](https://github.com/asaidimu/tasker/commit/c819e840cc8d22a60dd2c13e4ba2efbaca6ca528))

# 1.0.0 (2025-06-14)


* feat(core)!: introduce concurrent task management library ([1909fd6](https://github.com/asaidimu/tasker/commit/1909fd6156202bbcab20bfccb80aeb24c2719454))


### BREAKING CHANGES

* The previous placeholder module github.com/asaidimu/tasker/pkg and its Greeting function have been removed. The library now resides directly under the root github.com/asaidimu/tasker package.
