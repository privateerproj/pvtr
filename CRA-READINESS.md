# CRA Readiness

This project voluntarily documents its security practices using the Security Slam
[CRA Readiness checklist](https://securityslam.com/library/cra-readiness), aligned
with the EU [Cyber Resilience Act](https://openssf.org/public-policy/eu-cyber-resilience-act/).

## Disclaimer

> _This project voluntarily documents its security practices._
> _This information is provided "as is", without warranties or guarantees._
> _The maintainers and contributors:_
>
> - _have no obligations under the EU CRA,_
> - _are not Manufacturers, Importers, or Economic Operators,_
> - _assume no financial, contractual, or legal liability,_
> - _and do not provide CRA compliance assurances._
>
> _Entities incorporating this software into commercial products remain solely responsible for regulatory compliance, risk assessment, and vulnerability management._

For more context, see the ORC WG [maintainer transparency FAQ](https://cra.orcwg.org/faq/maintainers/transparency/).

## Checklist

Last reviewed: 2026-10-01

| Item | Description | Link to artifact |
| --- | --- | --- |
| Cybersecurity and Vulnerability Management Policy | Covers secure development practices, risk handling, security contact, the vulnerability reporting, remediation, and disclosure process, and the support period and end-of-life process. | [Security Policy](https://github.com/privateerproj/.github/blob/main/.github/SECURITY.md): [secure development](https://github.com/privateerproj/.github/blob/main/.github/SECURITY.md#secure-development), [risk handling](https://github.com/privateerproj/.github/blob/main/.github/SECURITY.md#risk-handling), [security contact](https://github.com/privateerproj/.github/blob/main/.github/SECURITY.md#security-contact), [vulnerability process](https://github.com/privateerproj/.github/blob/main/.github/SECURITY.md#vulnerability-process), [end of life](https://github.com/privateerproj/.github/blob/main/.github/SECURITY.md#end-of-life) |
| Contributing Guidance | Contributing guide links to secure development practices. | [CONTRIBUTING.md](https://github.com/privateerproj/pvtr/blob/main/CONTRIBUTING.md) |
| Release Documentation | Release notes describe new functionality and security fixes. | [GitHub Releases](https://github.com/privateerproj/pvtr/releases) |
| Bug Reporting Guide | Process for reporting non-security bugs, separate from security reporting. | [Bug report form](https://github.com/privateerproj/pvtr/issues/new?template=bug_report.yml). The [issue chooser](https://github.com/privateerproj/pvtr/issues/new/choose) sends security reports to the security policy instead. |
| MFA Enforcement | MFA is enabled for all contributors and required for admins. | The privateerproj GitHub organization requires two-factor authentication for all members and outside collaborators, including admins. |
| Branch Protection | Branch protection is enabled on the default branch. | A ruleset on `main` blocks deletion and force pushes, and requires a pull request with one approval, code owner review, resolved review threads, and passing status checks. |
| License File | The repository contains a clear, OSI-approved license. | [Apache-2.0](https://github.com/privateerproj/pvtr/blob/main/LICENSE) |
| OSPS Baseline | The project meets OSPS Baseline Level 1 or higher. | [OpenSSF Best Practices: Baseline Level 1](https://www.bestpractices.dev/projects/15145/baseline-1) |
