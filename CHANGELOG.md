# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Add CORS support
- Support user file uploads to S3 [#114](https://github.com/rokwire/content-building-block/issues/114)
- Add POST /content-items API
- Separate S3 signed URL expiration environment variables
### Changed
- Generate file IDs for S3 file uploads
- Reintroduce file extensions for S3

## [1.9.0] - 2024-12-03
### Added
- Add API to get voice record by user ID [#111](https://github.com/rokwire/content-building-block/issues/111)

## [1.8.0] - 2024-06-26
### Added
- Expose TPS API for adding images [#106](https://github.com/rokwire/content-building-block/issues/106)

## [1.7.0] - 2024-06-26
### Added
- Remove user data [#104](https://github.com/rokwire/content-building-block/issues/104)

## [1.6.1] - 2024-05-03
### Fixed
- Fix upload image BBs API [#100](https://github.com/rokwire/content-building-block/issues/100)

## [1.6.0] - 2024-04-30
### Added
- Create "upload image" bbs API [#97](https://github.com/rokwire/content-building-block/issues/97)

## [1.5.2] - 2024-02-15
###
- Use redirect to S3 for file downloads (EDIT: Stream files from S3) [#94](https://github.com/rokwire/content-building-block/issues/94)

## [1.5.1] - 2024-02-01
### Fixed
- Fix host/base URL environment variables

## [1.5.0] - 2024-01-24
### Added
- Refactor and support arbitrary file types [#87](https://github.com/rokwire/content-building-block/issues/87)

## [1.4.0] - 2024-01-12
### Added
- Expose delete user audio record API [#91](https://github.com/rokwire/content-building-block/issues/91)

## [1.3.0] - 2024-01-11
### Added
- Add user audio record availabillity [#89](https://github.com/rokwire/content-building-block/issues/89)

## [1.2.4] - 2023-07-11
### Added
- Set logger [#69](https://github.com/rokwire/content-building-block/issues/69)
- Prepare for deployment in OpenShift [#84](https://github.com/rokwire/content-building-block/issues/84)
### Fixed
- Photo upload uses local storage [#79](https://github.com/rokwire/content-building-block/issues/79)
- Fix docs [#76](https://github.com/rokwire/content-building-block/issues/76)
- Оrg_id could be lost on update operation [#81](https://github.com/rokwire/content-building-block/issues/81)

## [1.2.3] - 2022-08-01
### Fixed
- Missing admin image permission [#74](https://github.com/rokwire/content-building-block/issues/74)

## [1.2.2] - 2022-07-25
### Fixed
- Unable to upload profile images to S3. The S3 profile picture acl has been changed from "authenticated-read" to "private" [#72](https://github.com/rokwire/content-building-block/issues/72)
### Changed
- Prepare the project to become open source [#62](https://github.com/rokwire/content-building-block/issues/62)

## [1.2.1] - 2022-07-07
### Fixed
- Fix content item categories [#67](https://github.com/rokwire/content-building-block/issues/67)

## [1.2.0] - 2022-06-10
### Changed
- Update README file [#64](https://github.com/rokwire/content-building-block/issues/64)
- Multi-tenancy [#55](https://github.com/rokwire/content-building-block/issues/55)

## [1.1.12] - 2021-05-13
### Fixed
- Fix broken decode of content item [#60](https://github.com/rokwire/content-building-block/issues/60)

## [1.1.11] - 2021-05-12
### Fixed
- Fix bad documentation for content item APIs [#56](https://github.com/rokwire/content-building-block/issues/56)

## [1.1.10] - 2021-04-29
### Fixed
- Fix content item API inconsistencies [#52](https://github.com/rokwire/content-building-block/issues/52)

## [1.1.9] - 2021-04-28
### Changed
- Update Core auth library to the latest version and repo [#49](https://github.com/rokwire/content-building-block/issues/49)

## [1.1.8] - 2021-04-28
### Fixed
- Fix broken content items API [#47](https://github.com/rokwire/content-building-block/issues/47)

## [1.1.7] - 2021-04-26
### Changed
- Update Swagger library due to security issue [#45](https://github.com/rokwire/content-building-block/issues/45)

## [1.1.6] - 2021-04-04
### Fixed
- Bad response after requesting POST /image [#41](https://github.com/rokwire/content-building-block/issues/41)

## [1.1.5] - 2021-04-04
### Added
- Additional fix for setting appropriate ACL permission [#38](https://github.com/rokwire/content-building-block/issues/38)

## [1.1.4] - 2022-03-30
### Added
- Introduce CRUD of profile photos [#38](https://github.com/rokwire/content-building-block/issues/38)

## [1.1.3] - 2021-12-21
### Changed
- Introduce content item APIs (admin & client). More changes [#34](https://github.com/rokwire/content-building-block/issues/34)

## [1.1.2] - 2021-12-20
### Changed
- Introduce content item APIs (admin & client). More changes [#34](https://github.com/rokwire/content-building-block/issues/34)

## [1.1.1] - 2021-12-17
### Added
- Introduce content item APIs (admin & client) [#34](https://github.com/rokwire/content-building-block/issues/34)

## [1.0.14] - 2021-12-03
### Added
- Introduce & support health information data [#33](https://github.com/rokwire/content-building-block/issues/33)

## [1.0.13] - 2021-11-08
### Added
- Add support of all_admin_content permission [#30](https://github.com/rokwire/content-building-block/issues/30)

## [1.0.12] - 2021-11-05
### Fixed
- Update permissions and link core permissions [#28](https://github.com/rokwire/content-building-block/issues/28)

## [1.0.11] - 2021-10-25
### Fixed
- Support Core access tokens for UIUC client APIs [#27](https://github.com/rokwire/content-building-block/issues/27)

## [1.0.10] - 2021-10-08
### Fixed
- Twitter feed/widget just spins and spins [#25](https://github.com/rokwire/content-building-block/issues/25)

## [1.0.9] - 2021-10-05
### Added
- Integrate Core-BB [#22](https://github.com/rokwire/content-building-block/issues/22)

## [1.0.8] - 2021-09-17
### Fixed
- Additional fixes for image upload, cache expiration and readme improvements [#10](https://github.com/rokwire/content-building-block/issues/11)

## [1.0.7.3] - 2021-09-16
### Fixed
- Additional Twitter api improvements and fixes [#11](https://github.com/rokwire/content-building-block/issues/11)

## [1.0.7.2] - 2021-09-15
### Changed
- Fixed missing cwebp executable within the docker image

## [1.0.7.1] - 2021-09-15
### Changed 
- Fixed API doc & admin api security check[#10](https://github.com/rokwire/content-building-block/issues/10)

## [1.0.7] - 2021-09-15
### Added
- Implement upload image api [#10](https://github.com/rokwire/content-building-block/issues/10)
- Implement Twitter feed api [#11](https://github.com/rokwire/content-building-block/issues/11)
## [1.0.6] - 2021-06-23
### Fixed
- GET /student_guides results "null" response on missing data in the table [#8](https://github.com/rokwire/content-building-block/issues/8)
- Improve documentation [#6](https://github.com/rokwire/content-building-block/issues/6)

### Added
- Implement student guide client & admin APIs [#1](https://github.com/rokwire/content-building-block/issues/1)

