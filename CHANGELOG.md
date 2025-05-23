# Changelog

## [0.4.1](https://github.com/canonical/user-verification-service/compare/v0.4.0...v0.4.1) (2025-05-23)


### Bug Fixes

* add noop client ([89c5d03](https://github.com/canonical/user-verification-service/commit/89c5d036bcc0eb06cf1a74b70f5fb2507485f90c))

## [0.4.0](https://github.com/canonical/user-verification-service/compare/v0.3.1...v0.4.0) (2025-05-22)


### Features

* move from directory API to salesforce ([735b7b6](https://github.com/canonical/user-verification-service/commit/735b7b6727f2e6d71a8e8395840394cfe979c5f8))


### Bug Fixes

* copy payload on success ([cf07103](https://github.com/canonical/user-verification-service/commit/cf071035161a7750f9fb663c9fb60ee22e72e1a9))
* **deps:** update go deps to v1.36.0 ([36b861d](https://github.com/canonical/user-verification-service/commit/36b861df45eeba04d9a0e984b839c92f40e48175))
* **deps:** update go deps to v1.36.0 (minor) ([ff54135](https://github.com/canonical/user-verification-service/commit/ff54135e7cfebf2b6fc743dbf8a8e1e173ee94d2))

## [0.3.1](https://github.com/canonical/user-verification-service/compare/v0.3.0...v0.3.1) (2025-04-29)


### Bug Fixes

* fix routing ([eded5b7](https://github.com/canonical/user-verification-service/commit/eded5b72618cca176b4f77b3ea04b00e13057ce4))

## [0.3.0](https://github.com/canonical/user-verification-service/compare/v0.2.0...v0.3.0) (2025-04-29)


### Features

* add directory API client implementation ([87a8c40](https://github.com/canonical/user-verification-service/commit/87a8c40a8f313d9abf9cd5f09512d90971ff73ca))


### Bug Fixes

* add prom metric for directory API responses ([5eec20b](https://github.com/canonical/user-verification-service/commit/5eec20b265635088c95f149099498bbc5acbacc3))
* **deps:** update module go.uber.org/mock to v0.5.2 ([0d76353](https://github.com/canonical/user-verification-service/commit/0d76353d504e572aa8481108012d4e58b6c6d16e))
* **deps:** update module go.uber.org/mock to v0.5.2 ([09f7668](https://github.com/canonical/user-verification-service/commit/09f76687162cd03c9cc1eee1e36b0bf212ae769d))
* increase write timeout ([c2b6bfc](https://github.com/canonical/user-verification-service/commit/c2b6bfc8a44751cab2db0b18130517647ccaeb20))
* serve UI under subpath ([3f4c831](https://github.com/canonical/user-verification-service/commit/3f4c831ab155253e1b4370361983970b611d3962))
* update error message ([7194926](https://github.com/canonical/user-verification-service/commit/7194926d3c0266c686d6748399ca0af6d0f5860b))
* use itoa instead of format ([64282b5](https://github.com/canonical/user-verification-service/commit/64282b5189a6264e9e35b2253df43c57fc9f7047))

## [0.2.0](https://github.com/canonical/user-verification-service/compare/v0.1.0...v0.2.0) (2025-04-14)


### Features

* add initial go project structure ([763b4a0](https://github.com/canonical/user-verification-service/commit/763b4a04802e3608990147008fa373fbd151e7bb))
* add verify API ([e4fd29f](https://github.com/canonical/user-verification-service/commit/e4fd29fd2cff4479619b7b674b2ec46cf8a1ad98))


### Bug Fixes

* add dummy directory API client ([582fb0e](https://github.com/canonical/user-verification-service/commit/582fb0ef9828fe88221ddc3abd32dc8b3bbc3037))
* add registration error handler ([f17b483](https://github.com/canonical/user-verification-service/commit/f17b4831fc696d67b3c994dbe85fc2c7fec01bf6))
* add status API ([1b2d5a4](https://github.com/canonical/user-verification-service/commit/1b2d5a4d3f6ff06ffcb62588855404ce23e2808a))
* **deps:** update module github.com/prometheus/client_golang to v1.22.0 ([8678af6](https://github.com/canonical/user-verification-service/commit/8678af68f4c975dea4062bf941735dfe0052df2f))
* **deps:** update module github.com/prometheus/client_golang to v1.22.0 ([0112ec8](https://github.com/canonical/user-verification-service/commit/0112ec869871b0dc65e168b4c644c13aa9321709))
* **deps:** update module go.uber.org/mock to v0.5.1 ([ca97000](https://github.com/canonical/user-verification-service/commit/ca9700020fca3438862891049f773271b60706d3))
* **deps:** update module go.uber.org/mock to v0.5.1 ([4700107](https://github.com/canonical/user-verification-service/commit/4700107005f97a4d6409c938dc9f0930a8b1a339))
* fix typo ([a1630b7](https://github.com/canonical/user-verification-service/commit/a1630b755e35100341b6d47567370c3f45a05c95))
* log email for failed attempts ([e815ade](https://github.com/canonical/user-verification-service/commit/e815ade18d7049d28ab8b9f75bea97a49d84a889))
