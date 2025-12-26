## [1.24.0](https://github.com/dimasma0305/gzcli/compare/v1.23.1...v1.24.0) (2025-12-26)

### Features

* rename docker-compose.yml to compose.yml and add project root folder to the compose template. ([258b25c](https://github.com/dimasma0305/gzcli/commit/258b25c102e813c4e06e896619130df665960202))

### Code Refactoring

* split docker-compose into modular files, add new environment variables, and update README max flag length. ([c17186b](https://github.com/dimasma0305/gzcli/commit/c17186bbacda15c82e2aeda6a0f1c16ebc430bb7))

## [1.23.1](https://github.com/dimasma0305/gzcli/compare/v1.23.0...v1.23.1) (2025-12-26)

### Code Refactoring

* Update slug generation to use hyphens as separators and remove underscores from allowed characters. ([5f25695](https://github.com/dimasma0305/gzcli/commit/5f2569510f30e5b0cd66e51bb275b44b0c2aa9d6))

## [1.23.0](https://github.com/dimasma0305/gzcli/compare/v1.22.0...v1.23.0) (2025-12-26)

### Features

* Add interactive CSV column mapping for team creation, refactor CSV parsing with a `TeamConfig` struct, and update team creation email warnings. ([644fc1a](https://github.com/dimasma0305/gzcli/commit/644fc1a1324700383737ce6c31b6533b1c2ecb63))
* Isolate API cookies per user, improve team creation error handling, update challenge slug regex, and track example template `dist` directories. ([ba82171](https://github.com/dimasma0305/gzcli/commit/ba821711795bb8eef9e95ed24fdda804c7cb09af))
* permit underscores in challenge slugs and update gzcli to v1.22.1 in manager Dockerfile ([7d94d6b](https://github.com/dimasma0305/gzcli/commit/7d94d6bf57a86204121ae9ff2fb1a84ccb549414))

## [1.22.0](https://github.com/dimasma0305/gzcli/compare/v1.21.0...v1.22.0) (2025-12-17)

### Features

* Parameterize Traefik service, router, and middleware names with `{{.Workspace}}` for workspace-specific routing. ([60845a3](https://github.com/dimasma0305/gzcli/commit/60845a317512735bea8813780a993a7fdaf9f936))

## [1.21.0](https://github.com/dimasma0305/gzcli/compare/v1.20.0...v1.21.0) (2025-12-17)

### Features

* template WORKDIR in docker-compose.yml for dynamic root folder ([50ccd37](https://github.com/dimasma0305/gzcli/commit/50ccd3745b55a8911952bfa399040909e90b27ad))

## [1.20.0](https://github.com/dimasma0305/gzcli/compare/v1.19.0...v1.20.0) (2025-12-17)

### Features

* Dynamically set `RootFolder` in `docker-compose.yml` and add a test to verify its presence. ([c46821e](https://github.com/dimasma0305/gzcli/commit/c46821e0a6e10f9750a451d6ebb57c8e062f5e89))

## [1.19.0](https://github.com/dimasma0305/gzcli/compare/v1.18.0...v1.19.0) (2025-12-17)

### Features

* remove bot service and its related files from ctf-template ([8dda905](https://github.com/dimasma0305/gzcli/commit/8dda90574a6be6c86ace518a63a2c3382ac7696c))

## [1.18.0](https://github.com/dimasma0305/gzcli/compare/v1.17.1...v1.18.0) (2025-12-17)

### Features

* Introduce `gzcli` manager service with Docker integration, update CTF template configurations, and enhance `team create` command with event and invite code options. ([e5e5b5f](https://github.com/dimasma0305/gzcli/commit/e5e5b5f9fdb7187834eed38e18f71e3764a72c59))

## [1.17.1](https://github.com/dimasma0305/gzcli/compare/v1.17.0...v1.17.1) (2025-12-10)

### Code Refactoring

* update script failure handling and improve test coverage ([13f6b5b](https://github.com/dimasma0305/gzcli/commit/13f6b5b5b6a5d3d703d4a07fa0e2796c41300aa9))

## [1.17.0](https://github.com/dimasma0305/gzcli/compare/v1.16.0...v1.17.0) (2025-12-10)

### Features

* enhance script execution error handling and reporting ([c69e024](https://github.com/dimasma0305/gzcli/commit/c69e0249f8fd56ee57bb7a6860a818a1474b560d))

## [1.16.0](https://github.com/dimasma0305/gzcli/compare/v1.15.1...v1.16.0) (2025-12-10)

### Features

* add duplicate challenge removal and enhance challenge handling ([e512229](https://github.com/dimasma0305/gzcli/commit/e5122298e288fceef0a32f71bdab67ad93a8670c))
* complete sync reliability improvements and enhance challenge handling ([8e9650e](https://github.com/dimasma0305/gzcli/commit/8e9650e625893b09809f07f69e3ebafda97974ef))

## [1.15.1](https://github.com/dimasma0305/gzcli/compare/v1.15.0...v1.15.1) (2025-11-25)

### Code Refactoring

* migrate UI from Bootstrap to Tailwind CSS and remove starter templates section ([e5f785c](https://github.com/dimasma0305/gzcli/commit/e5f785c1836eaf7da1f5093dd23d99e2f6aa2131))

## [1.15.0](https://github.com/dimasma0305/gzcli/compare/v1.14.0...v1.15.0) (2025-11-23)

### Features

* automatically stop running challenges on server shutdown ([61fa638](https://github.com/dimasma0305/gzcli/commit/61fa6384c2ef68336f010f012b04a8218905324c))
* make challenge cleanup asynchronous on shutdown ([1280df0](https://github.com/dimasma0305/gzcli/commit/1280df00fd09e3f979b027936bfed33f9907a084))

## [1.13.2](https://github.com/dimasma0305/gzcli/compare/v1.13.1...v1.13.2) (2025-11-23)

### Bug Fixes

* port didn't randomize ([17cb6d8](https://github.com/dimasma0305/gzcli/commit/17cb6d89652930eb19c3f0fd5c36b69e2e011c9b))

## [1.13.1](https://github.com/dimasma0305/gzcli/compare/v1.13.0...v1.13.1) (2025-11-23)

### Bug Fixes

* port didn't show in frontend ([d085278](https://github.com/dimasma0305/gzcli/commit/d085278b0f91b253ff9b2d5ffcdc6fb3f3801950))

## [1.13.0](https://github.com/dimasma0305/gzcli/compare/v1.12.0...v1.13.0) (2025-11-23)

### Features

* randomize ports for dockerfile challenges ([7310e87](https://github.com/dimasma0305/gzcli/commit/7310e877234997e17afde7edbe229b8de3ec25e7))
* randomize ports for dockerfile challenges ([35af664](https://github.com/dimasma0305/gzcli/commit/35af6641befac220c7ad0c587e9b8d6902d70ca9))
* randomize ports for dockerfile challenges ([773f5a9](https://github.com/dimasma0305/gzcli/commit/773f5a94c38dc3d551a16185f1836a3e76eaff28))
* randomize ports for dockerfile challenges ([cea22f6](https://github.com/dimasma0305/gzcli/commit/cea22f62939e666974f171a1094a53d6a3c1ffee))
* randomize ports for dockerfile challenges ([895aa6a](https://github.com/dimasma0305/gzcli/commit/895aa6a83b18b7bd4883c8125b3d1c74cfc6ab25))

## [1.12.0](https://github.com/dimasma0305/gzcli/compare/v1.11.3...v1.12.0) (2025-11-21)

### Features

* Refactor core logic for improved readability and maintainability ([2f3e968](https://github.com/dimasma0305/gzcli/commit/2f3e968fd309a80ecd29d21d81a596e3ff55a904))

## [1.11.3](https://github.com/dimasma0305/gzcli/compare/v1.11.2...v1.11.3) (2025-11-10)

### Bug Fixes

* enhance upload server validation and template handling ([44359c5](https://github.com/dimasma0305/gzcli/commit/44359c5836a2dca6102c9967137c9df47b6990e5))

## [1.11.2](https://github.com/dimasma0305/gzcli/compare/v1.11.1...v1.11.2) (2025-11-09)

### Code Refactoring

* remove local validation instructions from upload page ([89507d9](https://github.com/dimasma0305/gzcli/commit/89507d9d4bd6ad1da95671da8f13842e0ebeb473))

## [1.11.1](https://github.com/dimasma0305/gzcli/compare/v1.11.0...v1.11.1) (2025-11-09)

### Code Refactoring

* enhance permission handling in install script ([5acd436](https://github.com/dimasma0305/gzcli/commit/5acd436f11c7452fc867d097048f4c6f2ec59071))

## [1.11.0](https://github.com/dimasma0305/gzcli/compare/v1.10.1...v1.11.0) (2025-11-09)

### Features

* add upload server feature ([7452dcc](https://github.com/dimasma0305/gzcli/commit/7452dccc3a00b71cfc73d5359e97125f3322302f))

### Bug Fixes

*  directory creation logic to ensure writable permissions ([bd95cc9](https://github.com/dimasma0305/gzcli/commit/bd95cc9c365e3c1168333d5fc59f6600c95ed584))
* lint issue ([537936f](https://github.com/dimasma0305/gzcli/commit/537936f4563c8405122a232c8192c20f92ab3b0c))

## [1.10.1](https://github.com/dimasma0305/gzcli/compare/v1.10.0...v1.10.1) (2025-10-12)

### Bug Fixes

* windows path issue ([5dcd323](https://github.com/dimasma0305/gzcli/commit/5dcd323c7fe2775272ebfacbfc66f3e02778100d))

## [1.10.0](https://github.com/dimasma0305/gzcli/compare/v1.9.3...v1.10.0) (2025-10-12)

### Features

* add discord webhook feature ([bd83368](https://github.com/dimasma0305/gzcli/commit/bd833688481fd348e54d072e67f04dba7e1323a0))

### Bug Fixes

* Add environment variable loading and expansion for Docker Compose and Dockerfile ([57f702b](https://github.com/dimasma0305/gzcli/commit/57f702b3004f99f6f0f4371c1e455a549ae20b08))

## [1.9.3](https://github.com/dimasma0305/gzcli/compare/v1.9.2...v1.9.3) (2025-10-12)

### Bug Fixes

* update the challenge template ([ca43693](https://github.com/dimasma0305/gzcli/commit/ca436932740ba8c091e93668df55b51af8f241f5))

## [1.9.2](https://github.com/dimasma0305/gzcli/compare/v1.9.1...v1.9.2) (2025-10-12)

### Bug Fixes

* timpa cache ([4828556](https://github.com/dimasma0305/gzcli/commit/482855639b27b3e6f2ffd47c973366458d416391))

## [1.9.1](https://github.com/dimasma0305/gzcli/compare/v1.9.0...v1.9.1) (2025-10-12)

### Bug Fixes

* looping issue in the event cache ([ee2d6ed](https://github.com/dimasma0305/gzcli/commit/ee2d6ed30f0499acdb92d96d9ef833dbcae20dda))

## [1.9.0](https://github.com/dimasma0305/gzcli/compare/v1.8.0...v1.9.0) (2025-10-12)

### Features

* Enhance challenge processing and host cache initialization ([1005dbb](https://github.com/dimasma0305/gzcli/commit/1005dbb634054148983136ac5d2d8d586a65c98d))

## [1.8.0](https://github.com/dimasma0305/gzcli/compare/v1.7.1...v1.8.0) (2025-10-09)

### Features

* Enhance challenge synchronization and normalization ([06a4e11](https://github.com/dimasma0305/gzcli/commit/06a4e111fe0cdd0bcecd9f380e070afce440c7cf))

### Code Refactoring

* Update challenge handling to use config types and improve code structure ([eaeb258](https://github.com/dimasma0305/gzcli/commit/eaeb25852a46899ebe9391ebbd2c615e23b9505d))

## [1.7.1](https://github.com/dimasma0305/gzcli/compare/v1.7.0...v1.7.1) (2025-10-08)

### Bug Fixes

* Improve error handling and test logic ([493a8b2](https://github.com/dimasma0305/gzcli/commit/493a8b26e04dcc8feb59d0667e41da7da7ca7012))
* linting priblem ([464e67e](https://github.com/dimasma0305/gzcli/commit/464e67e7138bd57712cddfc8a54a41d5c7541dce))
* Normalize paths in embeddedFS for Windows compatibility ([ba14f50](https://github.com/dimasma0305/gzcli/commit/ba14f5061d3432cf1b3a34b12f55db0bf28e5704))
* Use 0600 permissions for test files ([a8ff1d5](https://github.com/dimasma0305/gzcli/commit/a8ff1d56e13eb62e33774d4b8a7d16028f6dfa0e))

### Code Refactoring

* Improve linting, testing, and error handling ([87f2510](https://github.com/dimasma0305/gzcli/commit/87f251082278b3629955feac93af5c8be65e77c4))
* Improve test coverage and fix bugs ([fd38de7](https://github.com/dimasma0305/gzcli/commit/fd38de7a6ae0324504fc60753a5da463f26ce84c))
* Improve test coverage and fix bugs ([019487f](https://github.com/dimasma0305/gzcli/commit/019487f61073f178fb4246e8c8c557005fe5c519))
* Reduce file permissions in tests ([74b4963](https://github.com/dimasma0305/gzcli/commit/74b49636754791fe126e3af2354cede07bc453f5))

## [1.7.0](https://github.com/dimasma0305/gzcli/compare/v1.6.0...v1.7.0) (2025-10-08)

### Features

* add Challenge Launcher Server and WebSocket support ([90dd69f](https://github.com/dimasma0305/gzcli/commit/90dd69f49ecab510aa9d59983db4189798d8fd1c))

### Bug Fixes

* Ensure directory is restored before cleanup in tests ([156cf24](https://github.com/dimasma0305/gzcli/commit/156cf249b2089bc8f18fd0398739c7c63ff61098))

## [1.6.0](https://github.com/dimasma0305/gzcli/compare/v1.5.2...v1.6.0) (2025-10-08)

### Features

* update documentation and remove deprecated files ([8654d4d](https://github.com/dimasma0305/gzcli/commit/8654d4dbc62e0f7206f8e90c0d5e126a2f425c7b))

### Bug Fixes

* clean up whitespace in development practices documentation ([af517be](https://github.com/dimasma0305/gzcli/commit/af517bed5f6242e93f4a6e8ff8e5f0d7ebfd355b))

## [1.5.2](https://github.com/dimasma0305/gzcli/compare/v1.5.1...v1.5.2) (2025-10-08)

### Code Refactoring

* improve test cleanup and path normalization for cross-platform compatibility ([29fa312](https://github.com/dimasma0305/gzcli/commit/29fa3120f4b69230686e606818333ae432f73eeb))

## [1.5.1](https://github.com/dimasma0305/gzcli/compare/v1.5.0...v1.5.1) (2025-10-08)

### Bug Fixes

* resolve symlink paths for event poster configuration ([a162dd2](https://github.com/dimasma0305/gzcli/commit/a162dd2360191ac6795709a32d992c342402a71a))

## [1.5.0](https://github.com/dimasma0305/gzcli/compare/v1.4.1...v1.5.0) (2025-10-08)

### Features

* enhance event command functionality with multi-event support ([59f1b2c](https://github.com/dimasma0305/gzcli/commit/59f1b2cded71b5d0dda51ba22deff88e10bc4b6f))
* enhance event management with multi-event support and command improvements ([e954f1d](https://github.com/dimasma0305/gzcli/commit/e954f1de00afd9f675aa6540f1a247812911742d))
* implement event creation command with structured directory setup ([c1d0628](https://github.com/dimasma0305/gzcli/commit/c1d0628ed3094ae1d8fe6836bc56bbabb257d1b6))

### Code Refactoring

* enhance shell configuration handling in install script ([eb93393](https://github.com/dimasma0305/gzcli/commit/eb93393e88ec16439cbc545349e59ff0f65e1718))

## [1.4.1](https://github.com/dimasma0305/gzcli/compare/v1.4.0...v1.4.1) (2025-10-08)

### Code Refactoring

* improve migration process and enhance error handling ([bc2a944](https://github.com/dimasma0305/gzcli/commit/bc2a944caeeb542997b29ff5719aa7f8fc27d573))
* streamline installation and uninstallation processes ([939b6ac](https://github.com/dimasma0305/gzcli/commit/939b6acf3c47bf1f4571e4142d00ce1e239e10fb))

## [1.4.0](https://github.com/dimasma0305/gzcli/compare/v1.3.0...v1.4.0) (2025-10-08)

### Features

* add multi-event management and enhance configuration structure ([5701ffe](https://github.com/dimasma0305/gzcli/commit/5701ffe3257f02d94f691f5aeaad8483e5d99a56))

### Documentation

* update documentation for clarity and consistency ([22ece4f](https://github.com/dimasma0305/gzcli/commit/22ece4f9b005d1a92e3486ca16a0ecb61f37176c))

## [1.3.0](https://github.com/dimasma0305/gzcli/compare/v1.2.0...v1.3.0) (2025-10-08)

### Features

* enhance file write handling and improve temporary directory cleanup ([7cf4f2d](https://github.com/dimasma0305/gzcli/commit/7cf4f2d07d9a282e507fd6140b3777c150656926))

## [1.2.0](https://github.com/dimasma0305/gzcli/compare/v1.1.3...v1.2.0) (2025-10-08)

### Features

* fix cyclo test ([0da1f56](https://github.com/dimasma0305/gzcli/commit/0da1f56751130dc53cec28593a736e81e430a162))

## [1.1.3](https://github.com/dimasma0305/gzcli/compare/v1.1.2...v1.1.3) (2025-10-07)

### Bug Fixes

* clean up whitespace in benchmark tests and documentation ([d557e25](https://github.com/dimasma0305/gzcli/commit/d557e25c2d803e40c46eaa9687b136386d997fd1))

### Code Refactoring

* optimize caching mechanism and enhance file system event filtering ([75a80ae](https://github.com/dimasma0305/gzcli/commit/75a80ae263f6509fd952bd6a46b58494fba93032))

## [1.1.2](https://github.com/dimasma0305/gzcli/compare/v1.1.1...v1.1.2) (2025-10-07)

### Code Refactoring

* update script execution comments for security linting ([723bf78](https://github.com/dimasma0305/gzcli/commit/723bf786bdfd372197bb85e8455f0a49f35e261b))

## [1.1.1](https://github.com/dimasma0305/gzcli/compare/v1.1.0...v1.1.1) (2025-10-07)

### Code Refactoring

* improve cross-platform compatibility and testing logic ([a2bb8c6](https://github.com/dimasma0305/gzcli/commit/a2bb8c6a271b3cbebdaee5ab97fe36e8c86ba208))

## [1.1.0](https://github.com/dimasma0305/gzcli/compare/v1.0.0...v1.1.0) (2025-10-07)

### Features

* enhance installation script and .goreleaser configuration ([f3ce9df](https://github.com/dimasma0305/gzcli/commit/f3ce9dfb982be1f6d110ba1d5e260669027af520))

## 1.0.0 (2025-10-07)

### Features

* enhance account login and registration with response validation and improve network failure simulation in tests ([320443f](https://github.com/dimasma0305/gzcli/commit/320443f04c96ca27b57e922e8fc0b5847390033f))
* update .gitignore, .goreleaser.yml, Makefile, and root.go to include versioning metadata and build artifacts ([c4cb62a](https://github.com/dimasma0305/gzcli/commit/c4cb62a14df57d25591d2a3679309b94b71869a8))
