# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.0.0 (2026-11-21)


### Features

* **auth:** add managed identity and CI workload federation ([#214](https://github.com/deploymenttheory/go-sdk-msteams/issues/214)) ([ee3d2de](https://github.com/deploymenttheory/go-sdk-msteams/commit/ee3d2de597c7a7be88166bf73118170defdf19de))
* **auth:** add portable certificate and client-secret authentication ([#209](https://github.com/deploymenttheory/go-sdk-msteams/issues/209)) ([da821be](https://github.com/deploymenttheory/go-sdk-msteams/commit/da821bed6d3ac981fe9468c07b0b57b35e5816d8))
* **auth:** implement delegated login and Azure credential integration ([#212](https://github.com/deploymenttheory/go-sdk-msteams/issues/212)) ([32b4486](https://github.com/deploymenttheory/go-sdk-msteams/commit/32b4486747a0f2f68411f9da8f10aab7fa31fb63))
* **auth:** implement resource-aware supplied tokens and refresh coordination ([#207](https://github.com/deploymenttheory/go-sdk-msteams/issues/207)) ([99a6043](https://github.com/deploymenttheory/go-sdk-msteams/commit/99a6043f87005f7d61337f178a7426c7e8840227))
* **auth:** map PowerShell connection and environment helpers to SDK sessions ([#218](https://github.com/deploymenttheory/go-sdk-msteams/issues/218)) ([11fea7e](https://github.com/deploymenttheory/go-sdk-msteams/commit/11fea7eba82529f662b4acbea060af32c347bea9))
* **auth:** report command capabilities and operation access failures ([#216](https://github.com/deploymenttheory/go-sdk-msteams/issues/216)) ([641499f](https://github.com/deploymenttheory/go-sdk-msteams/commit/641499f9d6bdd3745e210202cbc60225cb64109e))
* **base:** add reproducible MicrosoftTeams baseline discovery ([#199](https://github.com/deploymenttheory/go-sdk-msteams/issues/199)) ([fbba04a](https://github.com/deploymenttheory/go-sdk-msteams/commit/fbba04aa86cdb112a02aa37c57bdfe594a962456))
* **base:** maintain the capability and verification coverage ledger ([#204](https://github.com/deploymenttheory/go-sdk-msteams/issues/204)) ([3df58a9](https://github.com/deploymenttheory/go-sdk-msteams/commit/3df58a9fda790e45dee67317a23a5ab43ef0e65e))
* **core:** add beta team-operation status and resumable waits ([#242](https://github.com/deploymenttheory/go-sdk-msteams/issues/242)) ([ac7914a](https://github.com/deploymenttheory/go-sdk-msteams/commit/ac7914ad1a5a13ada8b1010806c029a6d4503c80))
* **core:** add deployment and migration services ([#245](https://github.com/deploymenttheory/go-sdk-msteams/issues/245)) ([ca41361](https://github.com/deploymenttheory/go-sdk-msteams/commit/ca41361b9c37c997d5bfc6569761eefc81417e8c))
* **core:** add shared-channel relationships and command compositions ([#243](https://github.com/deploymenttheory/go-sdk-msteams/issues/243)) ([b771b02](https://github.com/deploymenttheory/go-sdk-msteams/commit/b771b02d06632ed655c6f8929d34a7492e99eaeb))
* **core:** add targeting policies and hierarchy operations ([#244](https://github.com/deploymenttheory/go-sdk-msteams/issues/244)) ([1e040cd](https://github.com/deploymenttheory/go-sdk-msteams/commit/1e040cd5e06ad2f9af42d9b39657491cdeb0cc5c))
* **core:** implement channel CRUD and management compositions ([#240](https://github.com/deploymenttheory/go-sdk-msteams/issues/240)) ([ac15c2d](https://github.com/deploymenttheory/go-sdk-msteams/commit/ac15c2db20b1f5f5ef48ba0d908667924a9821d3))
* **core:** implement channel membership CRUD and user compositions ([#241](https://github.com/deploymenttheory/go-sdk-msteams/issues/241)) ([6ee7b06](https://github.com/deploymenttheory/go-sdk-msteams/commit/6ee7b06bb366d70b0a3dd938ba7fccfe448e4cc5))
* **core:** implement memberships and team user compositions ([#237](https://github.com/deploymenttheory/go-sdk-msteams/issues/237)) ([1f08eeb](https://github.com/deploymenttheory/go-sdk-msteams/commit/1f08eebf0faa9dac66c86a8b3d244915a23d2e27))
* **core:** implement team resources and management compositions ([#234](https://github.com/deploymenttheory/go-sdk-msteams/issues/234)) ([da05a19](https://github.com/deploymenttheory/go-sdk-msteams/commit/da05a19931a6fc8c62adfd3f3a44cf9b5db8887d))
* **core:** implement team template CRUD and local builders ([#233](https://github.com/deploymenttheory/go-sdk-msteams/issues/233)) ([bd7ac7d](https://github.com/deploymenttheory/go-sdk-msteams/commit/bd7ac7dd5a629d0942486efb085f41a2a5886518))
* **delta:** compare semantic contracts and map SDK impact ([#206](https://github.com/deploymenttheory/go-sdk-msteams/issues/206)) ([82490a5](https://github.com/deploymenttheory/go-sdk-msteams/commit/82490a524e5fec82d2c2d0f69720d6880560cae6))
* **fnd:** wire root client configuration and typed service composition ([#205](https://github.com/deploymenttheory/go-sdk-msteams/issues/205)) ([8220cb8](https://github.com/deploymenttheory/go-sdk-msteams/commit/8220cb8abb108b39b7b8d8d7f29bcd997732eda0))
* **http:** add backend-aware pagination and partial list results ([#215](https://github.com/deploymenttheory/go-sdk-msteams/issues/215)) ([f6da1ce](https://github.com/deploymenttheory/go-sdk-msteams/commit/f6da1ce14ecebe36189c1f882de64ee1de8ca9de))
* **http:** add bounded operation polling and propagation waits ([#217](https://github.com/deploymenttheory/go-sdk-msteams/issues/217)) ([ab98359](https://github.com/deploymenttheory/go-sdk-msteams/commit/ab983594035002f0a97b724b70849a77eea56ddb))
* **http:** add bounded stream transfers and partial batch outcomes ([#219](https://github.com/deploymenttheory/go-sdk-msteams/issues/219)) ([df32cf0](https://github.com/deploymenttheory/go-sdk-msteams/commit/df32cf0bcd05970841e7e2cf3f9f2e5a503248ea))
* **http:** add redacted logging, tracing and transport configuration ([#222](https://github.com/deploymenttheory/go-sdk-msteams/issues/222)) ([47a4d84](https://github.com/deploymenttheory/go-sdk-msteams/commit/47a4d847ddc9a27d6bea5b84f302c7c3a6351348))
* **http:** bound retries, token replacement and concurrency ([#213](https://github.com/deploymenttheory/go-sdk-msteams/issues/213)) ([53c4a1d](https://github.com/deploymenttheory/go-sdk-msteams/commit/53c4a1dfc29b6e13279e0a3242040dc933fc1711))
* **http:** implement backend routing and normalized errors ([#211](https://github.com/deploymenttheory/go-sdk-msteams/issues/211)) ([dc4f993](https://github.com/deploymenttheory/go-sdk-msteams/commit/dc4f993cfe3683fa16671023b8a9aae8f48e449e))
* **http:** validate resource-specific paging success status ([#236](https://github.com/deploymenttheory/go-sdk-msteams/issues/236)) ([ccce2fd](https://github.com/deploymenttheory/go-sdk-msteams/commit/ccce2fdc6286082b13a5cf9f4e3af1f96b74bb16))
* **pol:** add application meeting configuration operations ([#246](https://github.com/deploymenttheory/go-sdk-msteams/issues/246)) ([81930ad](https://github.com/deploymenttheory/go-sdk-msteams/commit/81930ada6c83161529e591260d46d45913e3d1a2))
* **pol:** add federation resource operations ([#247](https://github.com/deploymenttheory/go-sdk-msteams/issues/247)) ([5f7d301](https://github.com/deploymenttheory/go-sdk-msteams/commit/5f7d301950e41eed52f0191617dec7dec3b8151d))
* **pol:** add native Teams custom banner text CRUD ([#266](https://github.com/deploymenttheory/go-sdk-msteams/issues/266)) ([cbca0f6](https://github.com/deploymenttheory/go-sdk-msteams/commit/cbca0f6c4390554662bceefe894d633af5556cff))
* **pol:** implement application access policy CRUD and assignments ([#231](https://github.com/deploymenttheory/go-sdk-msteams/issues/231)) ([3cb4aac](https://github.com/deploymenttheory/go-sdk-msteams/commit/3cb4aac9e58b10a2895adbc312b4388ad9ec078f))
* **pol:** implement audio conferencing custom prompts ([#258](https://github.com/deploymenttheory/go-sdk-msteams/issues/258)) ([fb84fc9](https://github.com/deploymenttheory/go-sdk-msteams/commit/fb84fc9e1380e67060929534aaa73edec67243a0))
* **pol:** implement call hold CRUD and modern grants ([#260](https://github.com/deploymenttheory/go-sdk-msteams/issues/260)) ([f1bc75c](https://github.com/deploymenttheory/go-sdk-msteams/commit/f1bc75c3278f45348164a42a54edc5021170bf5e))
* **pol:** implement call park policy CRUD and grants ([#261](https://github.com/deploymenttheory/go-sdk-msteams/issues/261)) ([5a45c71](https://github.com/deploymenttheory/go-sdk-msteams/commit/5a45c71edd896b12c3df3aabd3e7534f42dff82c))
* **pol:** implement calling policy CRUD and assignments ([#262](https://github.com/deploymenttheory/go-sdk-msteams/issues/262)) ([ce909da](https://github.com/deploymenttheory/go-sdk-msteams/commit/ce909da59e6dfa5cb496ca5157e5b5c7f22c3094))
* **pol:** implement channels policy CRUD and modern grants ([#263](https://github.com/deploymenttheory/go-sdk-msteams/issues/263)) ([f87bcf7](https://github.com/deploymenttheory/go-sdk-msteams/commit/f87bcf70b05cb3a2261117a3e6fa42cfef96b052))
* **pol:** implement client configuration reads and updates ([#264](https://github.com/deploymenttheory/go-sdk-msteams/issues/264)) ([2af6cb1](https://github.com/deploymenttheory/go-sdk-msteams/commit/2af6cb11f0623e8475f87bf93150c5666d12d94e))
* **pol:** implement Cortana policy CRUD and grants ([#265](https://github.com/deploymenttheory/go-sdk-msteams/issues/265)) ([38dcbf7](https://github.com/deploymenttheory/go-sdk-msteams/commit/38dcbf75dc45d8289793fa7f461d12b41de0ac00))
* **pol:** implement native assignment migration operations ([#248](https://github.com/deploymenttheory/go-sdk-msteams/issues/248)) ([cb1cc1e](https://github.com/deploymenttheory/go-sdk-msteams/commit/cb1cc1e657489b2ed2adbe52efe403112aace3b6))
* **pol:** implement native meeting branding policies ([#249](https://github.com/deploymenttheory/go-sdk-msteams/issues/249)) ([25180f7](https://github.com/deploymenttheory/go-sdk-msteams/commit/25180f777e69a3c12a095c2c7acf5ce288df3442))
* **pol:** implement native online voicemail policies ([#250](https://github.com/deploymenttheory/go-sdk-msteams/issues/250)) ([06a6b3c](https://github.com/deploymenttheory/go-sdk-msteams/commit/06a6b3cfb1ba1cac58c22b8f7f6339648ba52197))
* **pol:** implement policy assignment services ([#252](https://github.com/deploymenttheory/go-sdk-msteams/issues/252)) ([65e8575](https://github.com/deploymenttheory/go-sdk-msteams/commit/65e857567b3b090d55e9766abee2f01c8b738d9b))
* **pol:** implement policy package services ([#253](https://github.com/deploymenttheory/go-sdk-msteams/issues/253)) ([c763c35](https://github.com/deploymenttheory/go-sdk-msteams/commit/c763c35b5c8117f132838d3549d7fe303b95a0a9))
* **pol:** implement privacy configuration service ([#254](https://github.com/deploymenttheory/go-sdk-msteams/issues/254)) ([b909e84](https://github.com/deploymenttheory/go-sdk-msteams/commit/b909e8433b37206aec87c74eac084062a6ebe802))
* **pol:** implement recording message CRUD and localization builders ([#255](https://github.com/deploymenttheory/go-sdk-msteams/issues/255)) ([e26657a](https://github.com/deploymenttheory/go-sdk-msteams/commit/e26657a68232de863acfb944cedfd55ccbaaba8d))
* **pol:** implement Teams app permission policy CRUD and assignments ([#256](https://github.com/deploymenttheory/go-sdk-msteams/issues/256)) ([2ae9925](https://github.com/deploymenttheory/go-sdk-msteams/commit/2ae9925d3b04dba9f2cb7e1e4fe045e531fcc92a))
* **pol:** implement Teams app setup policy CRUD and assignments ([#257](https://github.com/deploymenttheory/go-sdk-msteams/issues/257)) ([8b65d6b](https://github.com/deploymenttheory/go-sdk-msteams/commit/8b65d6b1ce98a16e11cf1bbc257703cb8734a1ee))
* **pol:** implement Teams audio conferencing policies ([#259](https://github.com/deploymenttheory/go-sdk-msteams/issues/259)) ([f73e59e](https://github.com/deploymenttheory/go-sdk-msteams/commit/f73e59e797fffe87e86693395ca255f41612167d))
* **pol:** implement voicemail validation configuration ([#251](https://github.com/deploymenttheory/go-sdk-msteams/issues/251)) ([f6868d4](https://github.com/deploymenttheory/go-sdk-msteams/commit/f6868d42eda8bf0a935db3e35afc26dbceb9674e))


### Bug Fixes

* **ci:** serialize release-please branch updates ([#221](https://github.com/deploymenttheory/go-sdk-msteams/issues/221)) ([08ea119](https://github.com/deploymenttheory/go-sdk-msteams/commit/08ea11970f3f70cd6e0e6f353837082268e05f09))

## [Unreleased]

### Added

- Added xyz [@your_username](https://github.com/your_username)

### Fixed

- Fixed zyx [@your_username](https://github.com/your_username)

## [1.1.0] - 2021-06-23

### Added

- Added x [@your_username](https://github.com/your_username)

### Changed

- Changed y [@your_username](https://github.com/your_username)

## [1.0.0] - 2021-06-20

### Added

- Inititated y [@your_username](https://github.com/your_username)
- Inititated z [@your_username](https://github.com/your_username)
